package transaction_sub_transaction_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/domain"
	"jww/internal/service/transaction_sub_transaction"
	"jww/internal/service/unit_of_work"
)

func newAppContext() domain.AppContext {
	var appCtx domain.AppContext
	appCtx.SetContext(context.Background())
	return appCtx
}

// fakeRepo is a hand-written in-memory transaction_sub_transaction.Repository.
// It records call counts so tests can prove validation short-circuits before a
// port is touched, and can be made to fail per-method.
type fakeRepo struct {
	byID map[uuid.UUID]domain.TransactionSubTransaction

	createCalls int
	updateCalls int
	deleteCalls int
	getCalls    int
	listCalls   int

	createErr error
	updateErr error
	deleteErr error
	getErr    error
	listErr   error
}

func newFakeRepo(subTransactions ...domain.TransactionSubTransaction) *fakeRepo {
	byID := make(map[uuid.UUID]domain.TransactionSubTransaction, len(subTransactions))
	for _, s := range subTransactions {
		byID[s.ID] = s
	}
	return &fakeRepo{byID: byID}
}

func (r *fakeRepo) Create(_ domain.AppContext, s domain.TransactionSubTransaction, _ ...unit_of_work.UnitOfWork) error {
	r.createCalls++
	if r.createErr != nil {
		return r.createErr
	}
	r.byID[s.ID] = s
	return nil
}

func (r *fakeRepo) Update(_ domain.AppContext, s domain.TransactionSubTransaction, _ ...unit_of_work.UnitOfWork) error {
	r.updateCalls++
	if r.updateErr != nil {
		return r.updateErr
	}
	r.byID[s.ID] = s
	return nil
}

func (r *fakeRepo) Delete(_ domain.AppContext, s domain.TransactionSubTransaction, _ ...unit_of_work.UnitOfWork) error {
	r.deleteCalls++
	if r.deleteErr != nil {
		return r.deleteErr
	}
	delete(r.byID, s.ID)
	return nil
}

func (r *fakeRepo) GetByID(_ domain.AppContext, id uuid.UUID, _ ...unit_of_work.UnitOfWork) (*domain.TransactionSubTransaction, error) {
	r.getCalls++
	if r.getErr != nil {
		return nil, r.getErr
	}
	if s, ok := r.byID[id]; ok {
		return &s, nil
	}
	return nil, nil
}

func (r *fakeRepo) ListByTransactionIDs(_ domain.AppContext, transactionIDs []int64, _ ...unit_of_work.UnitOfWork) ([]domain.TransactionSubTransaction, error) {
	r.listCalls++
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
	return out, nil
}

type fakeUoW struct {
	restored bool
	finished bool
}

func (u *fakeUoW) Restore(domain.AppContext) error { u.restored = true; return nil }
func (u *fakeUoW) Finish(domain.AppContext) error  { u.finished = true; return nil }
func (u *fakeUoW) GetValue() interface{}           { return nil }

type fakeUoWManager struct{ created []*fakeUoW }

func (m *fakeUoWManager) Create(domain.AppContext) (unit_of_work.UnitOfWork, error) {
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

func seedSubTransaction(txID int64, name string, amount domain.Money) domain.TransactionSubTransaction {
	return domain.TransactionSubTransaction{
		ID:            uuid.New(),
		TransactionID: txID,
		Name:          name,
		Amount:        amount,
		CreatedAt:     time.Now().UTC(),
	}
}

func TestService_Create(t *testing.T) {
	t.Run("persists a validated row and returns it", func(t *testing.T) {
		repo := newFakeRepo()
		svc := transaction_sub_transaction.NewService(repo, &fakeUoWManager{})

		got, err := svc.Create(newAppContext(), 42, "Milk", "dairy", domain.Money(349))

		require.NoError(t, err)
		require.NotNil(t, got)
		assert.NotEqual(t, uuid.Nil, got.ID, "the service assigns the identity")
		assert.Equal(t, int64(42), got.TransactionID)
		assert.Equal(t, "Milk", got.Name)
		assert.Equal(t, "dairy", got.Label)
		assert.Equal(t, domain.Money(349), got.Amount)
		assert.Equal(t, 1, repo.createCalls)
		_, stored := repo.byID[got.ID]
		assert.True(t, stored, "the created row must reach the repository")
	})

	t.Run("blank name is rejected before the repository is called", func(t *testing.T) {
		repo := newFakeRepo()
		svc := transaction_sub_transaction.NewService(repo, &fakeUoWManager{})

		_, err := svc.Create(newAppContext(), 42, "   ", "", domain.Money(100))

		require.Error(t, err)
		assert.Zero(t, repo.createCalls, "validation must short-circuit the port")
	})

	t.Run("zero amount is rejected before the repository is called", func(t *testing.T) {
		repo := newFakeRepo()
		svc := transaction_sub_transaction.NewService(repo, &fakeUoWManager{})

		_, err := svc.Create(newAppContext(), 42, "Milk", "", 0)

		require.Error(t, err)
		assert.Zero(t, repo.createCalls)
	})

	t.Run("negative discount amount is accepted", func(t *testing.T) {
		repo := newFakeRepo()
		svc := transaction_sub_transaction.NewService(repo, &fakeUoWManager{})

		got, err := svc.Create(newAppContext(), 42, "Coupon", "discount", domain.Money(-100))

		require.NoError(t, err)
		assert.Equal(t, domain.Money(-100), got.Amount)
		assert.Equal(t, 1, repo.createCalls)
	})

	t.Run("an empty label is legitimate free text", func(t *testing.T) {
		repo := newFakeRepo()
		svc := transaction_sub_transaction.NewService(repo, &fakeUoWManager{})

		got, err := svc.Create(newAppContext(), 7, "Bread", "", domain.Money(199))

		require.NoError(t, err)
		assert.Equal(t, "", got.Label)
	})

	t.Run("repository failure propagates", func(t *testing.T) {
		repo := newFakeRepo()
		repo.createErr = errors.New("insert failed")
		svc := transaction_sub_transaction.NewService(repo, &fakeUoWManager{})

		_, err := svc.Create(newAppContext(), 42, "Milk", "", domain.Money(100))

		require.ErrorIs(t, err, repo.createErr)
	})
}

func TestService_GetByID(t *testing.T) {
	t.Run("returns the stored row", func(t *testing.T) {
		existing := seedSubTransaction(9, "Eggs", domain.Money(499))
		repo := newFakeRepo(existing)
		svc := transaction_sub_transaction.NewService(repo, &fakeUoWManager{})

		got, err := svc.GetByID(newAppContext(), existing.ID)

		require.NoError(t, err)
		assert.Equal(t, existing.ID, got.ID)
		assert.Equal(t, "Eggs", got.Name)
	})

	t.Run("a missing row is a not-found domain error", func(t *testing.T) {
		repo := newFakeRepo()
		svc := transaction_sub_transaction.NewService(repo, &fakeUoWManager{})

		_, err := svc.GetByID(newAppContext(), uuid.New())

		require.Error(t, err)
		var domainErr domain.DomainError
		require.ErrorAs(t, err, &domainErr)
		assert.Equal(t, domain.ErrorKindNotFound, domainErr.Kind())
		var notFound *domain.TransactionSubTransactionNotFoundError
		assert.ErrorAs(t, err, &notFound)
	})

	t.Run("repository error propagates unchanged", func(t *testing.T) {
		repo := newFakeRepo()
		repo.getErr = errors.New("boom")
		svc := transaction_sub_transaction.NewService(repo, &fakeUoWManager{})

		_, err := svc.GetByID(newAppContext(), uuid.New())
		require.ErrorIs(t, err, repo.getErr)
	})
}

func TestService_Update(t *testing.T) {
	t.Run("validates, checks existence, and persists in one unit of work", func(t *testing.T) {
		existing := seedSubTransaction(3, "Coffee", domain.Money(250))
		repo := newFakeRepo(existing)
		manager := &fakeUoWManager{}
		svc := transaction_sub_transaction.NewService(repo, manager)

		updated := existing
		updated.Name = "Espresso"
		updated.Label = "drinks"
		updated.Amount = domain.Money(300)
		got, err := svc.Update(newAppContext(), updated)

		require.NoError(t, err)
		assert.Equal(t, "Espresso", got.Name)
		assert.Equal(t, domain.Money(300), got.Amount)
		assert.Equal(t, "Espresso", repo.byID[existing.ID].Name)
		require.NotNil(t, manager.last())
		assert.True(t, manager.last().finished, "the unit of work commits on success")
		assert.False(t, manager.last().restored)
	})

	t.Run("a missing row is a not-found error and nothing is written", func(t *testing.T) {
		repo := newFakeRepo()
		manager := &fakeUoWManager{}
		svc := transaction_sub_transaction.NewService(repo, manager)

		ghost := seedSubTransaction(3, "Ghost", domain.Money(100))
		_, err := svc.Update(newAppContext(), ghost)

		require.Error(t, err)
		var domainErr domain.DomainError
		require.ErrorAs(t, err, &domainErr)
		assert.Equal(t, domain.ErrorKindNotFound, domainErr.Kind())
		assert.Zero(t, repo.updateCalls)
		require.NotNil(t, manager.last())
		assert.True(t, manager.last().restored, "every error path restores the unit of work")
	})

	t.Run("an invalid update is rejected before existence is checked", func(t *testing.T) {
		repo := newFakeRepo()
		svc := transaction_sub_transaction.NewService(repo, &fakeUoWManager{})

		bad := seedSubTransaction(3, "   ", domain.Money(100))
		_, err := svc.Update(newAppContext(), bad)

		require.Error(t, err)
		assert.Zero(t, repo.getCalls)
		assert.Zero(t, repo.updateCalls)
	})
}

func TestService_Delete(t *testing.T) {
	t.Run("deletes an existing row", func(t *testing.T) {
		existing := seedSubTransaction(3, "ToRemove", domain.Money(100))
		repo := newFakeRepo(existing)
		manager := &fakeUoWManager{}
		svc := transaction_sub_transaction.NewService(repo, manager)

		require.NoError(t, svc.Delete(newAppContext(), existing))

		assert.Equal(t, 1, repo.deleteCalls)
		_, stillThere := repo.byID[existing.ID]
		assert.False(t, stillThere)
		require.NotNil(t, manager.last())
		assert.True(t, manager.last().finished)
	})

	t.Run("a missing row is a not-found error and nothing is deleted", func(t *testing.T) {
		repo := newFakeRepo()
		manager := &fakeUoWManager{}
		svc := transaction_sub_transaction.NewService(repo, manager)

		err := svc.Delete(newAppContext(), seedSubTransaction(3, "Ghost", domain.Money(100)))

		require.Error(t, err)
		var domainErr domain.DomainError
		require.ErrorAs(t, err, &domainErr)
		assert.Equal(t, domain.ErrorKindNotFound, domainErr.Kind())
		assert.Zero(t, repo.deleteCalls)
		require.NotNil(t, manager.last())
		assert.True(t, manager.last().restored)
	})
}

func TestService_ListByTransactionIDs(t *testing.T) {
	t.Run("returns only rows for the requested parents", func(t *testing.T) {
		repo := newFakeRepo(
			seedSubTransaction(1, "A", domain.Money(100)),
			seedSubTransaction(1, "B", domain.Money(200)),
			seedSubTransaction(2, "C", domain.Money(300)),
		)
		svc := transaction_sub_transaction.NewService(repo, &fakeUoWManager{})

		got, err := svc.ListByTransactionIDs(newAppContext(), []int64{1})

		require.NoError(t, err)
		assert.Len(t, got, 2)
		for _, s := range got {
			assert.Equal(t, int64(1), s.TransactionID)
		}
	})

	t.Run("an empty id list short-circuits without hitting the repository", func(t *testing.T) {
		repo := newFakeRepo(seedSubTransaction(1, "A", domain.Money(100)))
		svc := transaction_sub_transaction.NewService(repo, &fakeUoWManager{})

		got, err := svc.ListByTransactionIDs(newAppContext(), nil)

		require.NoError(t, err)
		assert.Empty(t, got)
		assert.Zero(t, repo.listCalls)
	})

	t.Run("repository error propagates", func(t *testing.T) {
		repo := newFakeRepo()
		repo.listErr = errors.New("query failed")
		svc := transaction_sub_transaction.NewService(repo, &fakeUoWManager{})

		_, err := svc.ListByTransactionIDs(newAppContext(), []int64{1})
		require.ErrorIs(t, err, repo.listErr)
	})
}
