package entitlements_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/coder/coder/v2/coderd/entitlements"
	"github.com/coder/coder/v2/codersdk"
	"github.com/coder/coder/v2/testutil"
)

func TestModify(t *testing.T) {
	t.Parallel()

	set := entitlements.New()
	require.False(t, set.Enabled(codersdk.FeatureMultipleOrganizations))

	set.Modify(func(entitlements *codersdk.Entitlements) {
		entitlements.Features[codersdk.FeatureMultipleOrganizations] = codersdk.Feature{
			Enabled:     true,
			Entitlement: codersdk.EntitlementEntitled,
		}
	})
	require.True(t, set.Enabled(codersdk.FeatureMultipleOrganizations))
}

func TestAllowRefresh(t *testing.T) {
	t.Parallel()

	now := time.Now()
	set := entitlements.New()
	set.Modify(func(entitlements *codersdk.Entitlements) {
		entitlements.RefreshedAt = now
	})

	ok, wait := set.AllowRefresh(now)
	require.False(t, ok)
	require.InDelta(t, time.Minute.Seconds(), wait.Seconds(), 5)

	set.Modify(func(entitlements *codersdk.Entitlements) {
		entitlements.RefreshedAt = now.Add(time.Minute * -2)
	})

	ok, wait = set.AllowRefresh(now)
	require.True(t, ok)
	require.Equal(t, time.Duration(0), wait)
}

func TestUpdate(t *testing.T) {
	t.Parallel()
	ctx := testutil.Context(t, testutil.WaitShort)

	set := entitlements.New()
	require.False(t, set.Enabled(codersdk.FeatureMultipleOrganizations))
	fetchStarted := make(chan struct{})
	firstDone := make(chan struct{})
	errCh := make(chan error, 2)
	go func() {
		err := set.Update(ctx, func(_ context.Context) (codersdk.Entitlements, error) {
			close(fetchStarted)
			select {
			case <-firstDone:
				// OK!
			case <-ctx.Done():
				t.Error("timeout")
				return codersdk.Entitlements{}, ctx.Err()
			}
			return codersdk.Entitlements{
				Features: map[codersdk.FeatureName]codersdk.Feature{
					codersdk.FeatureMultipleOrganizations: {
						Enabled: true,
					},
				},
			}, nil
		})
		errCh <- err
	}()
	testutil.TryReceive(ctx, t, fetchStarted)
	require.False(t, set.Enabled(codersdk.FeatureMultipleOrganizations))
	// start a second update while the first one is in progress
	go func() {
		err := set.Update(ctx, func(_ context.Context) (codersdk.Entitlements, error) {
			return codersdk.Entitlements{
				Features: map[codersdk.FeatureName]codersdk.Feature{
					codersdk.FeatureMultipleOrganizations: {
						Enabled: true,
					},
					codersdk.FeatureAppearance: {
						Enabled: true,
					},
				},
			}, nil
		})
		errCh <- err
	}()
	close(firstDone)
	err := testutil.TryReceive(ctx, t, errCh)
	require.NoError(t, err)
	err = testutil.TryReceive(ctx, t, errCh)
	require.NoError(t, err)
	require.True(t, set.Enabled(codersdk.FeatureMultipleOrganizations))
	require.True(t, set.Enabled(codersdk.FeatureAppearance))
}

func TestUpdate_LicenseRequiresTelemetry(t *testing.T) {
	t.Parallel()
	ctx := testutil.Context(t, testutil.WaitShort)
	set := entitlements.New()
	set.Modify(func(entitlements *codersdk.Entitlements) {
		entitlements.Errors = []string{"some error"}
		entitlements.Features[codersdk.FeatureAppearance] = codersdk.Feature{
			Enabled: true,
		}
	})
	err := set.Update(ctx, func(_ context.Context) (codersdk.Entitlements, error) {
		return codersdk.Entitlements{}, entitlements.ErrLicenseRequiresTelemetry
	})
	require.NoError(t, err)
	require.True(t, set.Enabled(codersdk.FeatureAppearance))
	require.Equal(t, []string{entitlements.ErrLicenseRequiresTelemetry.Error()}, set.Errors())
}

func TestSetAlwaysEntitled(t *testing.T) {
	t.Parallel()
	ctx := testutil.Context(t, testutil.WaitShort)

	set := entitlements.New()
	set.SetAlwaysEntitled(codersdk.FeatureMultipleExternalAuth)

	// Simulate an enterprise entitlements fetch that returns the
	// feature as not entitled with an error about external auth.
	err := set.Update(ctx, func(_ context.Context) (codersdk.Entitlements, error) {
		return codersdk.Entitlements{
			Features: map[codersdk.FeatureName]codersdk.Feature{
				codersdk.FeatureMultipleExternalAuth: {
					Entitlement: codersdk.EntitlementNotEntitled,
					Enabled:     false,
				},
			},
			Errors: []string{
				"You have multiple External Auth Providers configured but this is an Enterprise feature. Reduce to one.",
			},
			Warnings: []string{
				"some unrelated warning",
			},
		}, nil
	})
	require.NoError(t, err)

	// The feature should be force-enabled.
	require.True(t, set.Enabled(codersdk.FeatureMultipleExternalAuth))
	f, ok := set.Feature(codersdk.FeatureMultipleExternalAuth)
	require.True(t, ok)
	require.Equal(t, codersdk.EntitlementEntitled, f.Entitlement)

	// The external auth error should be stripped, but unrelated
	// warnings should be preserved.
	require.Empty(t, set.Errors())
	require.Equal(t, []string{"some unrelated warning"}, set.Warnings())
}

func TestSetAlwaysEntitled_NoAlwaysEntitle(t *testing.T) {
	t.Parallel()
	ctx := testutil.Context(t, testutil.WaitShort)

	// Without SetAlwaysEntitled, errors pass through unchanged.
	set := entitlements.New()
	err := set.Update(ctx, func(_ context.Context) (codersdk.Entitlements, error) {
		return codersdk.Entitlements{
			Features: map[codersdk.FeatureName]codersdk.Feature{
				codersdk.FeatureMultipleExternalAuth: {
					Entitlement: codersdk.EntitlementNotEntitled,
					Enabled:     false,
				},
			},
			Errors: []string{
				"You have multiple External Auth Providers configured but this is an Enterprise feature. Reduce to one.",
			},
		}, nil
	})
	require.NoError(t, err)
	require.False(t, set.Enabled(codersdk.FeatureMultipleExternalAuth))
	require.Len(t, set.Errors(), 1)
}
