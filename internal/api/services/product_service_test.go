package services

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"go-crud-demo/models"
)

// ---------- Fake storage ----------
// fakeProductStore implement ProductStore ด้วย map ใน memory
// ไม่ต้องต่อ MongoDB จริง → test เร็วและรันได้ทุกที่
// ตั้ง err ไว้เมื่ออยากจำลองว่า DB ล่ม (ทุก method จะคืน error นี้)
type fakeProductStore struct {
	data   map[string]models.Product
	nextID int
	err    error
}

func newFakeStore(seed ...models.Product) *fakeProductStore {
	f := &fakeProductStore{data: map[string]models.Product{}}
	for _, p := range seed {
		f.data[p.ID] = p
	}
	return f
}

func (f *fakeProductStore) Insert(_ context.Context, p models.Product) (models.Product, error) {
	if f.err != nil {
		return models.Product{}, f.err
	}
	f.nextID++
	p.ID = fmt.Sprintf("id-%d", f.nextID)
	f.data[p.ID] = p
	return p, nil
}

func (f *fakeProductStore) FindAll(_ context.Context) ([]models.Product, error) {
	if f.err != nil {
		return nil, f.err
	}
	list := []models.Product{}
	for _, p := range f.data {
		list = append(list, p)
	}
	return list, nil
}

func (f *fakeProductStore) FindByID(_ context.Context, id string) (models.Product, error) {
	if f.err != nil {
		return models.Product{}, f.err
	}
	p, ok := f.data[id]
	if !ok {
		return models.Product{}, models.ErrNotFound
	}
	return p, nil
}

func (f *fakeProductStore) Update(_ context.Context, p models.Product) error {
	if f.err != nil {
		return f.err
	}
	if _, ok := f.data[p.ID]; !ok {
		return models.ErrNotFound
	}
	f.data[p.ID] = p
	return nil
}

func (f *fakeProductStore) Delete(_ context.Context, id string) error {
	if f.err != nil {
		return f.err
	}
	if _, ok := f.data[id]; !ok {
		return models.ErrNotFound
	}
	delete(f.data, id)
	return nil
}

// ใช้จำลอง error จาก DB
var errDBDown = errors.New("db down")

var coffee = models.Product{ID: "p1", Name: "Coffee", Price: 65}

// ---------- Tests ----------

func TestCreateProduct(t *testing.T) {
	tests := []struct {
		name    string
		request models.ProductRequest
		dbErr   error
		want    models.Product
		wantErr error
	}{
		{
			name:    "success",
			request: models.ProductRequest{Name: "Coffee", Price: 65},
			want:    models.Product{ID: "id-1", Name: "Coffee", Price: 65},
		},
		{
			name:    "trims spaces in name",
			request: models.ProductRequest{Name: "  Latte  ", Price: 75},
			want:    models.Product{ID: "id-1", Name: "Latte", Price: 75},
		},
		{
			name:    "storage error is returned",
			request: models.ProductRequest{Name: "Coffee", Price: 65},
			dbErr:   errDBDown,
			wantErr: errDBDown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newFakeStore()
			store.err = tt.dbErr
			svc := NewProductService(store) // ส่ง fake เข้าไปแทน storage จริง ผ่าน constructor

			got, err := svc.CreateProduct(context.Background(), tt.request)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
			if tt.wantErr == nil && store.data[got.ID] != tt.want {
				t.Errorf("stored %+v, want %+v", store.data[got.ID], tt.want)
			}
		})
	}
}

func TestGetAllProducts(t *testing.T) {
	tests := []struct {
		name    string
		seed    []models.Product
		dbErr   error
		wantLen int
		wantErr error
	}{
		{name: "empty", wantLen: 0},
		{name: "one product", seed: []models.Product{coffee}, wantLen: 1},
		{name: "storage error", dbErr: errDBDown, wantErr: errDBDown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newFakeStore(tt.seed...)
			store.err = tt.dbErr
			svc := NewProductService(store)

			got, err := svc.GetAllProducts(context.Background())

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if len(got) != tt.wantLen {
				t.Errorf("len = %d, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestGetProductByID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		dbErr   error
		want    models.Product
		wantErr error
	}{
		{name: "found", id: "p1", want: coffee},
		{name: "not found", id: "nope", wantErr: models.ErrNotFound},
		{name: "storage error", id: "p1", dbErr: errDBDown, wantErr: errDBDown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newFakeStore(coffee)
			store.err = tt.dbErr
			svc := NewProductService(store)

			got, err := svc.GetProductByID(context.Background(), tt.id)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestUpdateProduct(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		request models.ProductRequest
		dbErr   error
		want    models.Product
		wantErr error
	}{
		{
			name:    "success",
			id:      "p1",
			request: models.ProductRequest{Name: " Latte ", Price: 75},
			want:    models.Product{ID: "p1", Name: "Latte", Price: 75},
		},
		{
			name:    "not found",
			id:      "nope",
			request: models.ProductRequest{Name: "Latte", Price: 75},
			wantErr: models.ErrNotFound,
		},
		{
			name:    "storage error",
			id:      "p1",
			request: models.ProductRequest{Name: "Latte", Price: 75},
			dbErr:   errDBDown,
			wantErr: errDBDown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newFakeStore(coffee)
			store.err = tt.dbErr
			svc := NewProductService(store)

			got, err := svc.UpdateProduct(context.Background(), tt.id, tt.request)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
			if tt.wantErr == nil && store.data[tt.id] != tt.want {
				t.Errorf("stored %+v, want %+v", store.data[tt.id], tt.want)
			}
		})
	}
}

func TestDeleteProduct(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		dbErr   error
		wantErr error
	}{
		{name: "success", id: "p1"},
		{name: "not found", id: "nope", wantErr: models.ErrNotFound},
		{name: "storage error", id: "p1", dbErr: errDBDown, wantErr: errDBDown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newFakeStore(coffee)
			store.err = tt.dbErr
			svc := NewProductService(store)

			err := svc.DeleteProduct(context.Background(), tt.id)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if _, stillThere := store.data[tt.id]; tt.wantErr == nil && stillThere {
				t.Error("product was not deleted")
			}
		})
	}
}
