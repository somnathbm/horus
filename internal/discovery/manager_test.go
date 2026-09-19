package discovery

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/somnathbm/horus/internal/target"
)

// fake discoverers
type mockDiscoverer struct {
	result DiscoveryResult
	err    error
}

func (md mockDiscoverer) Discover(ctx context.Context) (DiscoveryResult, error) {
	if err := ctx.Err(); err != nil {
		return DiscoveryResult{}, err
	}

	return md.result, md.err
}

func TestManager(t *testing.T) {
	discovererNoFailure := mockDiscoverer{
		result: DiscoveryResult{
			Targets: []target.Target{
				{ID: "svc123", Name: "auth-service", Type: target.HTTPTargetType, Spec: target.HTTPTargetSpec{
					URL:  "http://v1.api.example.com/svcs/svc123",
					Port: 8080,
				}},
			},
		},
	}

	discovererWithFailure := mockDiscoverer{
		result: DiscoveryResult{
			Targets: []target.Target{
				{ID: "db123", Name: "payment-db", Type: target.PostgresSQLTargetType, Spec: target.PostgresTargetSpec{
					Host:     "example.db.com",
					Username: "dbadmin",
					Port:     8080,
				}},
			},
			Failures: []DiscoveryFailure{
				{Item: "redis.json", Err: errors.New("target \"cache123\": invalid schema")},
			},
		},
	}

	discovererNoTarget := mockDiscoverer{
		result: DiscoveryResult{
			// Targets: []target.Target{},
			Failures: []DiscoveryFailure{
				{Item: "redis.json", Err: errors.New("target \"cache123\": invalid schema")},
			},
		},
	}

	discovererResultWithError := mockDiscoverer{
		result: DiscoveryResult{},
		err:    errors.New("fatal error"),
	}

	// tests
	tests := []struct {
		name        string
		discoverers []Discoverer
		wantResult  DiscoveryResult
		wantErr     bool
	}{
		{
			name:        "zero-discoverer",
			discoverers: []Discoverer{},
			wantResult:  DiscoveryResult{},
			wantErr:     false,
		},
		{
			name:        "one-discoverer",
			discoverers: []Discoverer{discovererNoFailure},
			wantResult: DiscoveryResult{
				Targets: []target.Target{
					{ID: "svc123", Name: "auth-service", Type: target.HTTPTargetType, Spec: target.HTTPTargetSpec{
						URL:  "http://v1.api.example.com/svcs/svc123",
						Port: 8080,
					}},
				},
			},
			wantErr: false,
		},
		{
			name:        "multi-discoverers",
			discoverers: []Discoverer{discovererNoFailure, discovererWithFailure},
			wantResult: DiscoveryResult{
				Targets: []target.Target{
					{ID: "svc123", Name: "auth-service", Type: target.HTTPTargetType, Spec: target.HTTPTargetSpec{
						URL:  "http://v1.api.example.com/svcs/svc123",
						Port: 8080,
					}},
					{ID: "db123", Name: "payment-db", Type: target.PostgresSQLTargetType, Spec: target.PostgresTargetSpec{
						Host:     "example.db.com",
						Username: "dbadmin",
						Port:     8080,
					}},
				},
				Failures: []DiscoveryFailure{
					{Item: "redis.json", Err: errors.New("target \"cache123\": invalid schema")},
				},
			},
			wantErr: false,
		},
		{
			name:        "zero-target",
			discoverers: []Discoverer{discovererNoTarget},
			wantResult: DiscoveryResult{
				// Targets: []target.Target{},
				Failures: []DiscoveryFailure{
					{Item: "redis.json", Err: errors.New("target \"cache123\": invalid schema")},
				},
			},
			wantErr: false,
		},
		{
			name:        "fatal-err-discoverers-target-no-failure",
			discoverers: []Discoverer{discovererNoFailure, discovererResultWithError},
			wantResult: DiscoveryResult{
				Targets: []target.Target{
					{ID: "svc123", Name: "auth-service", Type: target.HTTPTargetType, Spec: target.HTTPTargetSpec{
						URL:  "http://v1.api.example.com/svcs/svc123",
						Port: 8080,
					}},
				},
			},
			wantErr: true,
		},
		{
			name:        "fatal-err-discoverers-target-with-failure",
			discoverers: []Discoverer{discovererNoFailure, discovererResultWithError, discovererWithFailure},
			wantResult: DiscoveryResult{
				Targets: []target.Target{
					{ID: "svc123", Name: "auth-service", Type: target.HTTPTargetType, Spec: target.HTTPTargetSpec{
						URL:  "http://v1.api.example.com/svcs/svc123",
						Port: 8080,
					}},
					{ID: "db123", Name: "payment-db", Type: target.PostgresSQLTargetType, Spec: target.PostgresTargetSpec{
						Host:     "example.db.com",
						Username: "dbadmin",
						Port:     8080,
					}},
				},
				Failures: []DiscoveryFailure{
					{Item: "redis.json", Err: errors.New("target \"cache123\": invalid schema")},
				},
			},
			wantErr: true,
		},
	}

	// canceled context
	t.Run("cancelled-context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		dscvryManager := New(discovererNoFailure)
		_, err := dscvryManager.Discover(ctx)
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expecting cancelled context, got: %v", err)
		}
	})

	// range over tests
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dscvryManager := New(tc.discoverers...)
			gotDscvrResult, err := dscvryManager.Discover(t.Context())

			if (err != nil) != tc.wantErr {
				t.Errorf("expected %v, got %v", tc.wantErr, err)
			}

			// target/failure count assertions
			if len(tc.wantResult.Targets) != len(gotDscvrResult.Targets) {
				t.Errorf("expected target count to be: %v, got: %v", len(tc.wantResult.Targets), len(gotDscvrResult.Targets))
			}
			if len(tc.wantResult.Failures) != len(gotDscvrResult.Failures) {
				t.Errorf("expected failure count to be: %v, got: %v", len(tc.wantResult.Failures), len(gotDscvrResult.Failures))
			}

			// targets assertions
			if !reflect.DeepEqual(tc.wantResult.Targets, gotDscvrResult.Targets) {
				t.Errorf("expected targets: %v, got: %v", tc.wantResult.Targets, gotDscvrResult.Targets)
			}
		})
	}
}
