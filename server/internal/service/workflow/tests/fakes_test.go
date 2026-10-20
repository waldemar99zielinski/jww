package workflow_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"sort"
	"time"

	"github.com/google/uuid"

	"jww/internal/domain"
	currency_exchange_rate "jww/internal/service/currency_exchange_rate"
	"jww/internal/service/object_storage"
	"jww/internal/service/transaction"
	"jww/internal/service/transaction_attachment"
	"jww/internal/service/unit_of_work"
	"jww/internal/service/webauthn_credential"
)

// newAppContext builds a minimal AppContext for tests that aren't driven by
// an *http.Request, mirroring scripts/bootstrap's non-HTTP construction.
func newAppContext() domain.AppContext {
	var appCtx domain.AppContext
	appCtx.SetContext(context.Background())
	return appCtx
}

// fakeUoW is a hand-written unit_of_work.UnitOfWork that records whether
// Restore/Finish were called, so workflow tests can assert every error path
// rolls back and the happy path commits exactly once.
type fakeUoW struct {
	restored   bool
	finished   bool
	restoreErr error
	finishErr  error
}

func (u *fakeUoW) Restore(domain.AppContext) error {
	u.restored = true
	return u.restoreErr
}

func (u *fakeUoW) Finish(domain.AppContext) error {
	u.finished = true
	return u.finishErr
}

func (u *fakeUoW) GetValue() interface{} { return nil }

// fakeUoWManager hands out fakeUoW instances and remembers the last one
// created, for assertions.
type fakeUoWManager struct {
	createErr error
	created   []*fakeUoW
}

func (m *fakeUoWManager) Create(domain.AppContext) (unit_of_work.UnitOfWork, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	u := &fakeUoW{}
	m.created = append(m.created, u)
	return u, nil
}

func (m *fakeUoWManager) last() *fakeUoW {
	if len(m.created) == 0 {
		return nil
	}
	return m.created[len(m.created)-1]
}

// fakeCurrencyRepo implements currency.Repository.
type fakeCurrencyRepo struct {
	byCode map[string]domain.Currency
	getErr error
}

func newFakeCurrencyRepo(currencies ...domain.Currency) *fakeCurrencyRepo {
	byCode := make(map[string]domain.Currency, len(currencies))
	for _, c := range currencies {
		byCode[c.Code] = c
	}
	return &fakeCurrencyRepo{byCode: byCode}
}

func (r *fakeCurrencyRepo) Create(domain.AppContext, domain.Currency, ...unit_of_work.UnitOfWork) error {
	return nil
}

func (r *fakeCurrencyRepo) GetByCode(_ domain.AppContext, code string, _ ...unit_of_work.UnitOfWork) (*domain.Currency, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	if c, ok := r.byCode[code]; ok {
		return &c, nil
	}
	return nil, nil
}

func (r *fakeCurrencyRepo) List(domain.AppContext, ...unit_of_work.UnitOfWork) ([]domain.Currency, error) {
	all := make([]domain.Currency, 0, len(r.byCode))
	for _, c := range r.byCode {
		all = append(all, c)
	}
	return all, nil
}

func (r *fakeCurrencyRepo) Update(domain.AppContext, domain.Currency, ...unit_of_work.UnitOfWork) error {
	return nil
}

func (r *fakeCurrencyRepo) Delete(domain.AppContext, domain.Currency, ...unit_of_work.UnitOfWork) error {
	return nil
}

// fakeExchangeRateRepo implements currency_exchange_rate.Repository.
type fakeExchangeRateRepo struct {
	byPair    map[string]domain.CurrencyExchangeRate // key: "FROM->TO"
	seeded    []domain.CurrencyExchangeRate
	getErr    error
	getCalls  int
	createErr error
	created   []domain.CurrencyExchangeRate
	listErr   error

	upsertErr    error
	upsertCalls  int
	upserted     []domain.CurrencyExchangeRateFlatten
	upsertCounts currency_exchange_rate.UpsertCounts
}

func newFakeExchangeRateRepo(rates ...domain.CurrencyExchangeRate) *fakeExchangeRateRepo {
	byPair := make(map[string]domain.CurrencyExchangeRate, len(rates))
	for _, r := range rates {
		byPair[r.FromCurrency.Code+"->"+r.ToCurrency.Code] = r
	}
	return &fakeExchangeRateRepo{byPair: byPair, seeded: rates}
}

func (r *fakeExchangeRateRepo) Create(_ domain.AppContext, e domain.CurrencyExchangeRate, _ ...unit_of_work.UnitOfWork) error {
	if r.createErr != nil {
		return r.createErr
	}
	r.created = append(r.created, e)
	r.byPair[e.FromCurrency.Code+"->"+e.ToCurrency.Code] = e
	return nil
}

func (r *fakeExchangeRateRepo) UpsertMany(_ domain.AppContext, rates []domain.CurrencyExchangeRateFlatten, _ ...unit_of_work.UnitOfWork) (currency_exchange_rate.UpsertCounts, error) {
	r.upsertCalls++
	if r.upsertErr != nil {
		return currency_exchange_rate.UpsertCounts{}, r.upsertErr
	}
	r.upserted = append(r.upserted, rates...)
	return r.upsertCounts, nil
}

func (r *fakeExchangeRateRepo) GetByCurrencies(_ domain.AppContext, from, to domain.Currency, _ time.Time, _ ...unit_of_work.UnitOfWork) (*domain.CurrencyExchangeRate, error) {
	r.getCalls++
	if r.getErr != nil {
		return nil, r.getErr
	}
	if e, ok := r.byPair[from.Code+"->"+to.Code]; ok {
		return &e, nil
	}
	return nil, nil
}

func (r *fakeExchangeRateRepo) List(domain.AppContext, *currency_exchange_rate.Filter, ...unit_of_work.UnitOfWork) ([]domain.CurrencyExchangeRate, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	all := make([]domain.CurrencyExchangeRate, 0, len(r.seeded)+len(r.created))
	all = append(all, r.seeded...)
	all = append(all, r.created...)
	return all, nil
}

// fakeRateSource implements currency_exchange_rate.RateSource. It records how
// it was called so tests can assert the codes/date the workflow asked for.
type fakeRateSource struct {
	fetch     currency_exchange_rate.RateFetch
	err       error
	calls     int
	lastCodes []string
	lastAsOf  time.Time
}

func (s *fakeRateSource) Rates(_ domain.AppContext, currencies []string, asOf time.Time) (currency_exchange_rate.RateFetch, error) {
	s.calls++
	s.lastCodes = currencies
	s.lastAsOf = asOf
	if s.err != nil {
		return currency_exchange_rate.RateFetch{}, s.err
	}
	return s.fetch, nil
}

// fakeCategoryRepo implements transactioncategory.Repository.
type fakeCategoryRepo struct {
	byID            map[uuid.UUID]domain.TransactionCategory
	getErr          error
	listByTypeCalls int
}

func newFakeCategoryRepo(categories ...domain.TransactionCategory) *fakeCategoryRepo {
	byID := make(map[uuid.UUID]domain.TransactionCategory, len(categories))
	for _, c := range categories {
		byID[c.ID] = c
	}
	return &fakeCategoryRepo{byID: byID}
}

func (r *fakeCategoryRepo) Create(domain.AppContext, domain.TransactionCategory, ...unit_of_work.UnitOfWork) error {
	return nil
}

func (r *fakeCategoryRepo) GetByID(_ domain.AppContext, id uuid.UUID, _ ...unit_of_work.UnitOfWork) (*domain.TransactionCategory, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	if c, ok := r.byID[id]; ok {
		return &c, nil
	}
	return nil, nil
}

func (r *fakeCategoryRepo) ListByType(domain.AppContext, domain.TransactionCategoryType, bool, ...unit_of_work.UnitOfWork) ([]domain.TransactionCategory, error) {
	r.listByTypeCalls++
	return nil, nil
}

func (r *fakeCategoryRepo) Update(domain.AppContext, domain.TransactionCategory, ...unit_of_work.UnitOfWork) error {
	return nil
}

func (r *fakeCategoryRepo) Delete(domain.AppContext, domain.TransactionCategory, ...unit_of_work.UnitOfWork) error {
	return nil
}

// fakeUserRepo implements user.Repository.
type fakeUserRepo struct {
	byID   map[uuid.UUID]domain.User
	getErr error
}

func newFakeUserRepo(users ...domain.User) *fakeUserRepo {
	byID := make(map[uuid.UUID]domain.User, len(users))
	for _, u := range users {
		byID[u.ID] = u
	}
	return &fakeUserRepo{byID: byID}
}

func (r *fakeUserRepo) Create(domain.AppContext, domain.User, ...unit_of_work.UnitOfWork) error {
	return nil
}

func (r *fakeUserRepo) GetByID(_ domain.AppContext, id uuid.UUID, _ ...unit_of_work.UnitOfWork) (*domain.User, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	if u, ok := r.byID[id]; ok {
		return &u, nil
	}
	return nil, nil
}

func (r *fakeUserRepo) List(domain.AppContext, ...unit_of_work.UnitOfWork) ([]domain.User, error) {
	return nil, nil
}

func (r *fakeUserRepo) Update(domain.AppContext, domain.User, ...unit_of_work.UnitOfWork) error {
	return nil
}

func (r *fakeUserRepo) Delete(domain.AppContext, domain.User, ...unit_of_work.UnitOfWork) error {
	return nil
}

// fakePasswordHasher implements user.PasswordHasher without doing any real
// hashing — this package only needs it to satisfy user.NewService.
type fakePasswordHasher struct{ hashErr error }

func (h fakePasswordHasher) Hash(password string) (string, error) {
	if h.hashErr != nil {
		return "", h.hashErr
	}
	return "hashed:" + password, nil
}

func (h fakePasswordHasher) Verify(password, hash string) bool {
	return hash == "hashed:"+password
}

// fakeTransactionRepo implements transaction.Repository.
type fakeTransactionRepo struct {
	createErr    error
	createCalled int
	nextID       int64

	listResult []domain.Transaction
	listErr    error
	listCalled int
	listFilter *transaction.Filter

	getResult *domain.Transaction
	getErr    error
	deleteErr error

	updated   []domain.Transaction
	deleted   []domain.Transaction
	deleteUOW unit_of_work.UnitOfWork
}

func (r *fakeTransactionRepo) Create(
	_ domain.AppContext,
	transactionType domain.TransactionType,
	amount domain.Money,
	currency domain.Currency,
	category domain.TransactionCategory,
	user domain.User,
	source domain.TransactionSource,
	createdAt time.Time,
	name string,
	description string,
	_ ...unit_of_work.UnitOfWork,
) (*domain.Transaction, error) {
	r.createCalled++
	if r.createErr != nil {
		return nil, r.createErr
	}
	r.nextID++
	return &domain.Transaction{
		ID:          r.nextID,
		Type:        transactionType,
		Amount:      amount,
		Currency:    currency,
		Category:    category,
		User:        user,
		Source:      source,
		Name:        name,
		Description: description,
		CreatedAt:   createdAt,
	}, nil
}

func (r *fakeTransactionRepo) GetByID(_ domain.AppContext, _ int64, _ ...unit_of_work.UnitOfWork) (*domain.Transaction, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	return r.getResult, nil
}

func (r *fakeTransactionRepo) List(_ domain.AppContext, filter *transaction.Filter, _ ...unit_of_work.UnitOfWork) ([]domain.Transaction, error) {
	r.listCalled++
	if filter != nil {
		copy := *filter
		r.listFilter = &copy
	}
	if r.listErr != nil {
		return nil, r.listErr
	}
	return r.listResult, nil
}

func (r *fakeTransactionRepo) Update(_ domain.AppContext, t domain.Transaction, _ ...unit_of_work.UnitOfWork) error {
	r.updated = append(r.updated, t)
	return nil
}

func (r *fakeTransactionRepo) Delete(_ domain.AppContext, t domain.Transaction, uow ...unit_of_work.UnitOfWork) error {
	if len(uow) > 0 {
		r.deleteUOW = uow[0]
	}
	if r.deleteErr != nil {
		return r.deleteErr
	}
	r.deleted = append(r.deleted, t)
	return nil
}

// fakeAttachmentRepo implements transaction_attachment.Repository in memory.
type fakeAttachmentRepo struct {
	byID       map[uuid.UUID]domain.TransactionAttachment
	createErr  error
	linkErr    error
	listErr    error
	created    []domain.TransactionAttachment
	linked     []domain.TransactionAttachment
	linkCalls  int
	deleteUOWs []unit_of_work.UnitOfWork
}

func newFakeAttachmentRepo(attachments ...domain.TransactionAttachment) *fakeAttachmentRepo {
	byID := make(map[uuid.UUID]domain.TransactionAttachment, len(attachments))
	for _, a := range attachments {
		byID[a.ID] = a
	}
	return &fakeAttachmentRepo{byID: byID}
}

func (r *fakeAttachmentRepo) Create(_ domain.AppContext, a domain.TransactionAttachment, _ ...unit_of_work.UnitOfWork) error {
	if r.createErr != nil {
		return r.createErr
	}
	r.created = append(r.created, a)
	r.byID[a.ID] = a
	return nil
}

func (r *fakeAttachmentRepo) GetByID(_ domain.AppContext, id uuid.UUID, _ ...unit_of_work.UnitOfWork) (*domain.TransactionAttachment, error) {
	a, ok := r.byID[id]
	if !ok {
		return nil, nil
	}
	return &a, nil
}

func (r *fakeAttachmentRepo) Find(_ domain.AppContext, filter transaction_attachment.Filter, _ ...unit_of_work.UnitOfWork) (*domain.TransactionAttachment, error) {
	a, ok := r.byID[filter.ID]
	if !ok || filter.DraftOnly && a.TransactionID != nil {
		return nil, nil
	}
	if filter.TransactionID != nil && (a.TransactionID == nil || *a.TransactionID != *filter.TransactionID) {
		return nil, nil
	}
	return &a, nil
}

func (r *fakeAttachmentRepo) LockByIDs(_ domain.AppContext, ids []uuid.UUID, _ ...unit_of_work.UnitOfWork) ([]domain.TransactionAttachment, error) {
	var out []domain.TransactionAttachment
	for _, id := range ids {
		if a, ok := r.byID[id]; ok {
			out = append(out, a)
		}
	}
	return out, nil
}

func (r *fakeAttachmentRepo) LinkToTransaction(_ domain.AppContext, transactionID int64, ids []uuid.UUID, _ ...unit_of_work.UnitOfWork) error {
	r.linkCalls++
	if r.linkErr != nil {
		return r.linkErr
	}
	for _, id := range ids {
		a, ok := r.byID[id]
		if !ok || a.TransactionID != nil {
			continue
		}
		txID := transactionID
		a.TransactionID = &txID
		r.byID[id] = a
		r.linked = append(r.linked, a)
	}
	return nil
}

func (r *fakeAttachmentRepo) ListByTransactionIDs(_ domain.AppContext, ids []int64, _ ...unit_of_work.UnitOfWork) ([]domain.TransactionAttachment, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	want := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		want[id] = struct{}{}
	}
	var out []domain.TransactionAttachment
	for _, a := range r.byID {
		if a.TransactionID == nil {
			continue
		}
		if _, ok := want[*a.TransactionID]; ok {
			out = append(out, a)
		}
	}
	return out, nil
}

func (r *fakeAttachmentRepo) Delete(_ domain.AppContext, a domain.TransactionAttachment, uow ...unit_of_work.UnitOfWork) error {
	if len(uow) > 0 {
		r.deleteUOWs = append(r.deleteUOWs, uow[0])
	}
	delete(r.byID, a.ID)
	return nil
}

// fakeAttachmentStorage implements object_storage.Store.
type fakeAttachmentStorage struct {
	objects   map[string][]byte
	putErr    error
	openErr   error
	deleteErr error
	putCalls  int
	deleted   []string
}

func newFakeAttachmentStorage() *fakeAttachmentStorage {
	return &fakeAttachmentStorage{objects: map[string][]byte{}}
}

func (s *fakeAttachmentStorage) Put(_ domain.AppContext, ref object_storage.Ref, body io.Reader, _ string) (object_storage.Stored, error) {
	s.putCalls++
	if s.putErr != nil {
		return object_storage.Stored{}, s.putErr
	}
	b, err := io.ReadAll(body)
	if err != nil {
		return object_storage.Stored{}, err
	}
	s.objects[ref.Key] = b
	return object_storage.Stored{Version: "v1", ETag: "etag-v1"}, nil
}

func (s *fakeAttachmentStorage) Open(_ domain.AppContext, ref object_storage.Ref) (object_storage.Stream, error) {
	if s.openErr != nil {
		return object_storage.Stream{}, s.openErr
	}
	b, ok := s.objects[ref.Key]
	if !ok {
		return object_storage.Stream{}, object_storage.ErrNotFound
	}
	return object_storage.Stream{Body: io.NopCloser(bytes.NewReader(b)), ByteSize: int64(len(b))}, nil
}

func (s *fakeAttachmentStorage) OpenRange(_ domain.AppContext, ref object_storage.Ref, byteRange object_storage.ByteRange) (object_storage.Stream, error) {
	b, ok := s.objects[ref.Key]
	if !ok {
		return object_storage.Stream{}, object_storage.ErrNotFound
	}
	b = b[byteRange.Start : byteRange.End+1]
	return object_storage.Stream{Body: io.NopCloser(bytes.NewReader(b)), ByteSize: int64(len(b))}, nil
}

func (s *fakeAttachmentStorage) Delete(_ domain.AppContext, ref object_storage.Ref) error {
	if s.deleteErr != nil {
		return s.deleteErr
	}
	s.deleted = append(s.deleted, ref.Key)
	delete(s.objects, ref.Key)
	return nil
}

// fakeSubTransactionRepo implements transaction_sub_transaction.Repository in
// memory, mirroring the parent-order the Postgres adapter uses (created_at,
// then id).
type fakeSubTransactionRepo struct {
	byID      map[uuid.UUID]domain.TransactionSubTransaction
	createErr error
	updateErr error
	deleteErr error
	getErr    error
	listErr   error
}

func newFakeSubTransactionRepo(subs ...domain.TransactionSubTransaction) *fakeSubTransactionRepo {
	byID := make(map[uuid.UUID]domain.TransactionSubTransaction, len(subs))
	for _, s := range subs {
		byID[s.ID] = s
	}
	return &fakeSubTransactionRepo{byID: byID}
}

func (r *fakeSubTransactionRepo) Create(_ domain.AppContext, s domain.TransactionSubTransaction, _ ...unit_of_work.UnitOfWork) error {
	if r.createErr != nil {
		return r.createErr
	}
	r.byID[s.ID] = s
	return nil
}

func (r *fakeSubTransactionRepo) Update(_ domain.AppContext, s domain.TransactionSubTransaction, _ ...unit_of_work.UnitOfWork) error {
	if r.updateErr != nil {
		return r.updateErr
	}
	r.byID[s.ID] = s
	return nil
}

func (r *fakeSubTransactionRepo) Delete(_ domain.AppContext, s domain.TransactionSubTransaction, _ ...unit_of_work.UnitOfWork) error {
	if r.deleteErr != nil {
		return r.deleteErr
	}
	delete(r.byID, s.ID)
	return nil
}

func (r *fakeSubTransactionRepo) GetByID(_ domain.AppContext, id uuid.UUID, _ ...unit_of_work.UnitOfWork) (*domain.TransactionSubTransaction, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	if s, ok := r.byID[id]; ok {
		return &s, nil
	}
	return nil, nil
}

func (r *fakeSubTransactionRepo) ListByTransactionIDs(_ domain.AppContext, transactionIDs []int64, _ ...unit_of_work.UnitOfWork) ([]domain.TransactionSubTransaction, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	want := make(map[int64]struct{}, len(transactionIDs))
	for _, id := range transactionIDs {
		want[id] = struct{}{}
	}
	var out []domain.TransactionSubTransaction
	for _, s := range r.byID {
		if _, ok := want[s.TransactionID]; ok {
			out = append(out, s)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if !out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].CreatedAt.Before(out[j].CreatedAt)
		}
		return out[i].ID.String() < out[j].ID.String()
	})
	return out, nil
}

// fakeWebAuthnCredentialRepo implements webauthn_credential.Repository.
type fakeWebAuthnCredentialRepo struct {
	byID           map[uuid.UUID]domain.WebAuthnCredential
	byCredentialID map[string]uuid.UUID
	createErr      error
	getByIDErr     error
	getByCredErr   error
	getByUserErr   error
	updateSignErr  error
	deleteErr      error
}

func newFakeWebAuthnCredentialRepo(credentials ...domain.WebAuthnCredential) *fakeWebAuthnCredentialRepo {
	byID := make(map[uuid.UUID]domain.WebAuthnCredential, len(credentials))
	byCredentialID := make(map[string]uuid.UUID, len(credentials))
	for _, c := range credentials {
		byID[c.ID] = c
		byCredentialID[string(c.CredentialID)] = c.ID
	}
	return &fakeWebAuthnCredentialRepo{byID: byID, byCredentialID: byCredentialID}
}

func (r *fakeWebAuthnCredentialRepo) Create(_ domain.AppContext, c domain.WebAuthnCredential, _ ...unit_of_work.UnitOfWork) error {
	if r.createErr != nil {
		return r.createErr
	}
	r.byID[c.ID] = c
	r.byCredentialID[string(c.CredentialID)] = c.ID
	return nil
}

func (r *fakeWebAuthnCredentialRepo) GetByID(_ domain.AppContext, id uuid.UUID, _ ...unit_of_work.UnitOfWork) (*domain.WebAuthnCredential, error) {
	if r.getByIDErr != nil {
		return nil, r.getByIDErr
	}
	if c, ok := r.byID[id]; ok {
		return &c, nil
	}
	return nil, nil
}

func (r *fakeWebAuthnCredentialRepo) GetByCredentialID(_ domain.AppContext, credentialID []byte, _ ...unit_of_work.UnitOfWork) (*domain.WebAuthnCredential, error) {
	if r.getByCredErr != nil {
		return nil, r.getByCredErr
	}
	if id, ok := r.byCredentialID[string(credentialID)]; ok {
		c := r.byID[id]
		return &c, nil
	}
	return nil, nil
}

func (r *fakeWebAuthnCredentialRepo) GetByUserID(_ domain.AppContext, userID uuid.UUID, _ ...unit_of_work.UnitOfWork) ([]domain.WebAuthnCredential, error) {
	if r.getByUserErr != nil {
		return nil, r.getByUserErr
	}
	var out []domain.WebAuthnCredential
	for _, c := range r.byID {
		if c.UserID == userID {
			out = append(out, c)
		}
	}
	return out, nil
}

func (r *fakeWebAuthnCredentialRepo) UpdateSignCount(_ domain.AppContext, id uuid.UUID, signCount uint32, lastUsedAt time.Time, _ ...unit_of_work.UnitOfWork) error {
	if r.updateSignErr != nil {
		return r.updateSignErr
	}
	c := r.byID[id]
	c.SignCount = signCount
	c.LastUsedAt = &lastUsedAt
	r.byID[id] = c
	return nil
}

func (r *fakeWebAuthnCredentialRepo) Delete(_ domain.AppContext, id uuid.UUID, _ ...unit_of_work.UnitOfWork) error {
	if r.deleteErr != nil {
		return r.deleteErr
	}
	delete(r.byID, id)
	return nil
}

// fakeCeremony implements webauthn_credential.Ceremony.
type fakeCeremony struct {
	beginRegOptions domain.WebAuthnCeremonyOptions
	beginRegSession domain.WebAuthnCeremonySession
	beginRegErr     error

	finishRegCredentialID, finishRegPublicKey, finishRegAAGUID []byte
	finishRegAttestationType                                   string
	finishRegSignCount                                         uint32
	finishRegTransports                                        []string
	finishRegBackupEligible                                    bool
	finishRegErr                                               error

	beginLoginOptions domain.WebAuthnCeremonyOptions
	beginLoginSession domain.WebAuthnCeremonySession
	beginLoginErr     error

	// loginRawID/loginUserHandle are passed to the handler FinishLogin
	// invokes, simulating what an authenticator response would carry.
	loginRawID, loginUserHandle []byte
	// postHandlerErr simulates the ceremony's own cryptographic
	// verification failing even after the handler successfully resolved a
	// user (e.g. bad signature, clone-counter mismatch).
	postHandlerErr          error
	finishLoginCredentialID []byte
	finishLoginSignCount    uint32
}

func (c *fakeCeremony) BeginRegistration(domain.User, []domain.WebAuthnCredential) (domain.WebAuthnCeremonyOptions, domain.WebAuthnCeremonySession, error) {
	return c.beginRegOptions, c.beginRegSession, c.beginRegErr
}

func (c *fakeCeremony) FinishRegistration(domain.User, []domain.WebAuthnCredential, domain.WebAuthnCeremonySession, domain.WebAuthnCeremonyResponse) ([]byte, []byte, []byte, string, uint32, []string, bool, error) {
	return c.finishRegCredentialID, c.finishRegPublicKey, c.finishRegAAGUID, c.finishRegAttestationType, c.finishRegSignCount, c.finishRegTransports, c.finishRegBackupEligible, c.finishRegErr
}

func (c *fakeCeremony) BeginLogin() (domain.WebAuthnCeremonyOptions, domain.WebAuthnCeremonySession, error) {
	return c.beginLoginOptions, c.beginLoginSession, c.beginLoginErr
}

func (c *fakeCeremony) FinishLogin(handler webauthn_credential.DiscoverableUserHandler, _ domain.WebAuthnCeremonySession, _ domain.WebAuthnCeremonyResponse) (domain.User, []byte, uint32, error) {
	u, _, err := handler(c.loginRawID, c.loginUserHandle)
	if err != nil {
		return domain.User{}, nil, 0, err
	}
	if c.postHandlerErr != nil {
		return domain.User{}, nil, 0, c.postHandlerErr
	}
	return u, c.finishLoginCredentialID, c.finishLoginSignCount, nil
}

// fakeChallengeTokens implements webauthn_credential.ChallengeTokens.
type fakeChallengeTokens struct {
	issueRegToken    string
	issueRegErr      error
	verifyRegUserID  uuid.UUID
	verifyRegSession json.RawMessage
	verifyRegErr     error

	issueLoginToken    string
	issueLoginErr      error
	verifyLoginSession json.RawMessage
	verifyLoginErr     error
}

func (c *fakeChallengeTokens) IssueRegistration(uuid.UUID, json.RawMessage) (string, error) {
	return c.issueRegToken, c.issueRegErr
}

func (c *fakeChallengeTokens) VerifyRegistration(string) (uuid.UUID, json.RawMessage, error) {
	return c.verifyRegUserID, c.verifyRegSession, c.verifyRegErr
}

func (c *fakeChallengeTokens) IssueLogin(json.RawMessage) (string, error) {
	return c.issueLoginToken, c.issueLoginErr
}

func (c *fakeChallengeTokens) VerifyLogin(string) (json.RawMessage, error) {
	return c.verifyLoginSession, c.verifyLoginErr
}

// fakeTokenService implements auth.TokenService.
type fakeTokenService struct {
	issueToken string
	issueErr   error
}

func (s *fakeTokenService) Issue(uuid.UUID) (string, error) {
	return s.issueToken, s.issueErr
}

func (s *fakeTokenService) Verify(string) (*uuid.UUID, error) {
	return nil, nil
}
