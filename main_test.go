package main

import "testing"

func TestValidateServiceTenant(t *testing.T) {
	tests := []struct {
		name             string
		authRequired     bool
		endpointOverride string
		endpointTenantID string
		wantErr          bool
	}{
		{name: "service requires concrete tenant", authRequired: true, wantErr: true},
		{name: "service accepts concrete tenant", authRequired: true, endpointTenantID: "00000000-0000-0000-0000-000000000000"},
		{name: "service accepts explicit endpoint", authRequired: true, endpointOverride: "http://127.0.0.1:8080/"},
		{name: "non-service command does not require tenant"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateServiceTenant(tt.authRequired, tt.endpointOverride, tt.endpointTenantID)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateServiceTenant() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
