// Package bootstrap creates a complete, realistic, repeatable local
// development data set.
package bootstrap

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"io/fs"
	"math/rand"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"jww/internal/adapter/gateway"
	"jww/internal/adapter/security"
	"jww/internal/domain"
	"jww/internal/infrastructure/database"
	"jww/internal/service/currency"
	currency_exchange_rate "jww/internal/service/currency_exchange_rate"
	"jww/internal/service/object_storage"
	"jww/internal/service/transaction"
	"jww/internal/service/transaction_attachment"
	"jww/internal/service/transaction_category"
	"jww/internal/service/user"
)

const markerPrefix = "[bootstrap:v1:"

//go:embed assets/*
var fixtureAssets embed.FS

// Options controls the amount and repeatability of generated data.
type Options struct {
	TransactionCount int
	AttachmentCount  int
	RandomSeed       int64
	Now              time.Time
}

func DefaultOptions() Options {
	return Options{
		TransactionCount: 5_000,
		AttachmentCount:  75,
		RandomSeed:       20_261_025,
		Now:              time.Now().UTC(),
	}
}

func (o Options) validate() error {
	if o.TransactionCount < 0 {
		return fmt.Errorf("transaction count must not be negative")
	}
	if o.AttachmentCount < 0 {
		return fmt.Errorf("attachment count must not be negative")
	}
	if o.AttachmentCount > o.TransactionCount {
		return fmt.Errorf("attachment count (%d) must not exceed transaction count (%d)", o.AttachmentCount, o.TransactionCount)
	}
	if o.Now.IsZero() {
		return fmt.Errorf("current time must not be zero")
	}
	return nil
}

// Result summarizes changes made by one run. TransactionsFound is the number
// of bootstrap-owned transactions found before missing records were created.
type Result struct {
	RateDays          int
	RatesInserted     int
	RatesUpdated      int
	TransactionsFound int
	TransactionsAdded int
	AttachmentsAdded  int
}

type services struct {
	users        *user.Service
	currencies   *currency.Service
	categories   *transaction_category.Service
	transactions *transaction.Service
	attachments  *transaction_attachment.Service
	rates        *currency_exchange_rate.Service
	uow          *database.PostgresUnitOfWorkManager
}

// Seed bootstraps fixed users/currencies and then creates rates, transactions,
// and attachments through the same repositories and services as the server.
// The object store may be nil only when AttachmentCount is zero.
func Seed(ctx context.Context, pool *pgxpool.Pool, store object_storage.Store, opts Options) (Result, error) {
	if err := opts.validate(); err != nil {
		return Result{}, err
	}
	if opts.AttachmentCount > 0 && store == nil {
		return Result{}, fmt.Errorf("object storage is required when attachments are requested")
	}
	if err := seedReferences(ctx, pool); err != nil {
		return Result{}, fmt.Errorf("bootstrapping users and currencies: %w", err)
	}

	appCtx := newAppContext(ctx)
	s := newServices(pool, store)
	users, currencies, expenseCategories, incomeCategories, err := loadReferences(appCtx, s)
	if err != nil {
		return Result{}, err
	}

	result, err := seedRates(appCtx, s.rates, currencies, opts.Now)
	if err != nil {
		return result, err
	}

	allTransactions, err := s.transactions.List(appCtx, nil)
	if err != nil {
		return result, fmt.Errorf("listing existing transactions: %w", err)
	}
	owned := make(map[int]domain.Transaction, opts.TransactionCount)
	for _, tr := range allTransactions {
		if index, ok := markerIndex(tr.Description); ok && index < opts.TransactionCount {
			owned[index] = tr
		}
	}
	result.TransactionsFound = len(owned)

	plans := transactionPlans(opts, users, currencies, expenseCategories, incomeCategories)
	if len(owned) < opts.TransactionCount {
		uow, err := s.uow.Create(appCtx)
		if err != nil {
			return result, fmt.Errorf("starting transaction seed: %w", err)
		}
		for i, plan := range plans {
			if _, exists := owned[i]; exists {
				continue
			}
			created, err := s.transactions.Create(
				appCtx, plan.Type, plan.Amount, plan.Currency, plan.Category,
				plan.User, plan.Source, plan.Name, plan.Description, &plan.CreatedAt, uow,
			)
			if err != nil {
				_ = uow.Restore(appCtx)
				return result, fmt.Errorf("creating transaction fixture %d: %w", i, err)
			}
			owned[i] = *created
			result.TransactionsAdded++
		}
		if err := uow.Finish(appCtx); err != nil {
			return result, fmt.Errorf("committing transaction seed: %w", err)
		}
	}

	added, err := seedAttachments(appCtx, s.attachments, owned, opts)
	result.AttachmentsAdded = added
	if err != nil {
		return result, err
	}
	return result, nil
}

func newServices(pool *pgxpool.Pool, store object_storage.Store) services {
	uow := database.NewPostgresUnitOfWorkManager(pool)
	userRepo := gateway.NewPostgresUserRepository(pool)
	currencyRepo := gateway.NewPostgresCurrencyRepository(pool)
	categoryRepo := gateway.NewPostgresTransactionCategoryRepository(pool)
	transactionRepo := gateway.NewPostgresTransactionRepository(pool)
	attachmentRepo := gateway.NewPostgresTransactionAttachmentRepository(pool)

	return services{
		users:        user.NewService(userRepo, security.NewArgon2PasswordHasher(), uow),
		currencies:   currency.NewService(currencyRepo, uow),
		categories:   transaction_category.NewService(categoryRepo, uow),
		transactions: transaction.NewService(transactionRepo, uow),
		attachments:  transaction_attachment.NewService(attachmentRepo, store, nil),
		rates:        currency_exchange_rate.NewService(gateway.NewPostgresCurrencyExchangeRateRepository(pool)),
		uow:          uow,
	}
}

func loadReferences(ctx domain.AppContext, s services) ([]domain.User, []domain.Currency, []domain.TransactionCategory, []domain.TransactionCategory, error) {
	users, err := s.users.List(ctx)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("listing users: %w", err)
	}
	currencies, err := s.currencies.List(ctx)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("listing currencies: %w", err)
	}
	expenses, err := s.categories.ListByType(ctx, domain.TransactionCategoryTypeExpense, true)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("listing expense categories: %w", err)
	}
	incomes, err := s.categories.ListByType(ctx, domain.TransactionCategoryTypeIncome, true)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("listing income categories: %w", err)
	}
	if len(users) == 0 || len(currencies) < 2 || len(expenses) == 0 || len(incomes) == 0 {
		return nil, nil, nil, nil, fmt.Errorf("development seed requires users, at least two currencies, and income/expense categories")
	}
	return users, currencies, expenses, incomes, nil
}

func seedRates(ctx domain.AppContext, service *currency_exchange_rate.Service, currencies []domain.Currency, now time.Time) (Result, error) {
	start, end := seedRange(now)
	rates := generatedRates(currencies, start, end)
	counts, err := service.UpsertMany(ctx, rates)
	result := Result{
		RateDays:      daysInclusive(start, end),
		RatesInserted: counts.Inserted,
		RatesUpdated:  counts.Updated,
	}
	if err != nil {
		return result, fmt.Errorf("seeding exchange rates: %w", err)
	}
	return result, nil
}

func generatedRates(currencies []domain.Currency, start, end time.Time) []domain.CurrencyExchangeRateFlatten {
	// PLN values represent the scaled PLN value of one unit. Small deterministic
	// daily movements make charts non-flat while keeping reciprocal pairs close.
	base := map[string]int64{"PLN": 10_000, "USD": 40_000, "EUR": 43_500}
	var rates []domain.CurrencyExchangeRateFlatten
	for day, dayIndex := start, int64(0); !day.After(end); day, dayIndex = day.AddDate(0, 0, 1), dayIndex+1 {
		values := map[string]int64{
			"PLN": base["PLN"],
			"USD": base["USD"] + ((dayIndex*137)%1_201 - 600),
			"EUR": base["EUR"] + ((dayIndex*173)%1_001 - 500),
		}
		for _, from := range currencies {
			fromValue, fromOK := values[from.Code]
			if !fromOK {
				continue
			}
			for _, to := range currencies {
				toValue, toOK := values[to.Code]
				if from.Code == to.Code || !toOK {
					continue
				}
				rate := (fromValue*10_000 + toValue/2) / toValue
				rates = append(rates, domain.CurrencyExchangeRateFlatten{
					FromCurrencyCode: from.Code,
					ToCurrencyCode:   to.Code,
					Rate:             domain.ExchangeRate(rate),
					RecordedAt:       day,
				})
			}
		}
	}
	return rates
}

type transactionPlan struct {
	Type        domain.TransactionType
	Amount      domain.Money
	Currency    domain.Currency
	Category    domain.TransactionCategory
	User        domain.User
	Source      domain.TransactionSource
	Name        string
	Description string
	CreatedAt   time.Time
}

func transactionPlans(opts Options, users []domain.User, currencies []domain.Currency, expenses, incomes []domain.TransactionCategory) []transactionPlan {
	rng := rand.New(rand.NewSource(opts.RandomSeed))
	start, _ := seedRange(opts.Now)
	spanSeconds := int64(opts.Now.UTC().Sub(start).Seconds())
	if spanSeconds < 1 {
		spanSeconds = 1
	}

	plans := make([]transactionPlan, 0, opts.TransactionCount)
	for i := 0; i < opts.TransactionCount; i++ {
		isIncome := rng.Intn(100) < 12
		categories := expenses
		transactionType := domain.TransactionTypeExpense
		if isIncome {
			categories = incomes
			transactionType = domain.TransactionTypeIncome
		}
		category := categories[rng.Intn(len(categories))]
		currency := weightedCurrency(rng, currencies)
		source := domain.TransactionSourceManual
		if rng.Intn(100) < 14 {
			source = domain.TransactionSourceRecurring
		}
		amount := expenseAmount(rng, category.Name)
		if isIncome {
			amount = domain.Money(25_000 + rng.Int63n(975_001))
		}
		createdAt := start.Add(time.Duration(rng.Int63n(spanSeconds)) * time.Second)
		plans = append(plans, transactionPlan{
			Type:        transactionType,
			Amount:      amount,
			Currency:    currency,
			Category:    category,
			User:        users[rng.Intn(len(users))],
			Source:      source,
			Name:        transactionName(rng, category.Name, isIncome),
			Description: fmt.Sprintf("Generated local development data %s%06d]", markerPrefix, i),
			CreatedAt:   createdAt,
		})
	}
	return plans
}

func weightedCurrency(rng *rand.Rand, currencies []domain.Currency) domain.Currency {
	byCode := make(map[string]domain.Currency, len(currencies))
	for _, c := range currencies {
		byCode[c.Code] = c
	}
	preferred := "PLN"
	roll := rng.Intn(100)
	if roll >= 82 && roll < 92 {
		preferred = "EUR"
	} else if roll >= 92 {
		preferred = "USD"
	}
	if c, ok := byCode[preferred]; ok {
		return c
	}
	return currencies[rng.Intn(len(currencies))]
}

func expenseAmount(rng *rand.Rand, category string) domain.Money {
	switch category {
	case "Rent":
		return domain.Money(180_000 + rng.Int63n(220_001))
	case "Travel":
		return domain.Money(8_000 + rng.Int63n(242_001))
	case "Utilities", "Education", "Health":
		return domain.Money(2_000 + rng.Int63n(78_001))
	default:
		return domain.Money(250 + rng.Int63n(39_751))
	}
}

var namesByCategory = map[string][]string{
	"Groceries":     {"Weekly groceries", "Corner market", "Fresh produce", "Household shopping"},
	"Transport":     {"Train ticket", "Fuel", "City transport", "Taxi ride"},
	"Entertainment": {"Cinema", "Concert tickets", "Board game", "Streaming rental"},
	"Health":        {"Pharmacy", "Dentist", "Health check", "Vitamins"},
	"Rent":          {"Monthly rent", "Apartment payment"},
	"Clothes":       {"Shoes", "Winter clothes", "Everyday clothing"},
	"Subscriptions": {"Music subscription", "Cloud storage", "Video subscription"},
	"Travel":        {"Hotel", "Flight tickets", "Weekend trip", "Travel booking"},
	"Utilities":     {"Electricity bill", "Internet bill", "Water bill", "Heating"},
	"Education":     {"Online course", "Books", "Language lessons"},
	"Personal care": {"Hairdresser", "Cosmetics", "Skincare"},
	"Gifts":         {"Birthday gift", "Wedding gift", "Holiday present"},
	"Dining out":    {"Lunch", "Dinner out", "Coffee and cake", "Takeaway"},
	"Baby":          {"Baby supplies", "Toys", "Children's clothes"},
	"Home":          {"Home supplies", "Furniture", "Small repairs", "Garden shop"},
	"Salary":        {"Monthly salary", "Payroll"},
	"Freelance":     {"Client invoice", "Consulting work", "Freelance project"},
	"Investments":   {"Dividend", "Interest payment", "Investment return"},
	"Refunds":       {"Shop refund", "Travel refund", "Returned purchase"},
	"Bonus":         {"Performance bonus", "Annual bonus"},
}

func transactionName(rng *rand.Rand, category string, income bool) string {
	if names := namesByCategory[category]; len(names) > 0 {
		return names[rng.Intn(len(names))]
	}
	if income {
		return category + " income"
	}
	return category + " expense"
}

func markerIndex(description string) (int, bool) {
	start := strings.Index(description, markerPrefix)
	if start < 0 {
		return 0, false
	}
	var index int
	var closing rune
	if _, err := fmt.Sscanf(description[start:], markerPrefix+"%d%c", &index, &closing); err != nil || closing != ']' || index < 0 {
		return 0, false
	}
	return index, true
}

type asset struct {
	Filename  string
	MediaType string
	Body      []byte
}

func loadAssets() ([]asset, error) {
	entries, err := fs.ReadDir(fixtureAssets, "assets")
	if err != nil {
		return nil, err
	}
	assets := make([]asset, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		body, err := fixtureAssets.ReadFile("assets/" + entry.Name())
		if err != nil {
			return nil, err
		}
		mediaType := "application/octet-stream"
		switch strings.ToLower(filepath.Ext(entry.Name())) {
		case ".pdf":
			mediaType = "application/pdf"
		case ".jpg", ".jpeg":
			mediaType = "image/jpeg"
		case ".png":
			mediaType = "image/png"
		}
		assets = append(assets, asset{Filename: entry.Name(), MediaType: mediaType, Body: body})
	}
	sort.Slice(assets, func(i, j int) bool { return assets[i].Filename < assets[j].Filename })
	if len(assets) == 0 {
		return nil, fmt.Errorf("no embedded development assets")
	}
	return assets, nil
}

func seedAttachments(ctx domain.AppContext, service *transaction_attachment.Service, transactions map[int]domain.Transaction, opts Options) (int, error) {
	if opts.AttachmentCount == 0 {
		return 0, nil
	}
	assets, err := loadAssets()
	if err != nil {
		return 0, fmt.Errorf("loading attachment fixtures: %w", err)
	}

	added := 0
	for n := 0; n < opts.AttachmentCount; n++ {
		index := n * opts.TransactionCount / opts.AttachmentCount
		tr, ok := transactions[index]
		if !ok {
			return added, fmt.Errorf("attachment target transaction fixture %d is missing", index)
		}
		existing, err := service.ListForTransaction(ctx, tr.ID)
		if err != nil {
			return added, fmt.Errorf("listing attachments for transaction %d: %w", tr.ID, err)
		}
		alreadySeeded := false
		for _, attachment := range existing {
			if strings.HasPrefix(attachment.OriginalFilename, "bootstrap-") {
				alreadySeeded = true
				break
			}
		}
		if alreadySeeded {
			continue
		}
		fixture := assets[n%len(assets)]
		if _, err := service.UploadForTransaction(ctx, tr.ID, fixture.Filename, fixture.MediaType, bytes.NewReader(fixture.Body)); err != nil {
			return added, fmt.Errorf("uploading attachment for transaction fixture %d: %w", index, err)
		}
		added++
	}
	return added, nil
}

func seedRange(now time.Time) (time.Time, time.Time) {
	now = now.UTC()
	end := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return end.AddDate(0, -2, 0), end
}

func daysInclusive(start, end time.Time) int {
	days := 0
	for day := start; !day.After(end); day = day.AddDate(0, 0, 1) {
		days++
	}
	return days
}
