package contextutil

import "context"

type tenantContextKey struct{}

type TenantContext struct {
	tenantID uint
}

func NewTenantContext(tenantID uint) *TenantContext {
	return &TenantContext{tenantID: tenantID}
}

func (tc *TenantContext) TenantID() uint {
	return tc.tenantID
}

func (tc *TenantContext) SetTenantID(id uint) {
	tc.tenantID = id
}

func GetTenantContext(ctx context.Context) *TenantContext {
	if tc, ok := ctx.Value(tenantContextKey{}).(*TenantContext); ok {
		return tc
	}
	return nil
}

func WithTenantContext(ctx context.Context, tc *TenantContext) context.Context {
	return context.WithValue(ctx, tenantContextKey{}, tc)
}
