package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	config "github.com/darialissi/avito_merch_service/lib/config"
	"github.com/darialissi/avito_merch_service/tests/testutil"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func getConfigPath(tb testing.TB) string {
	tb.Helper()

	if path := os.Getenv("CONFIG_PATH"); path != "" {
		return path
	}

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		tb.Fatalf("cannot resolve caller path")
		return ""
	}

	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", ".."))
	return filepath.Join(repoRoot, "configs", ".config.test.yml")
}

func mustCreatePool(tb testing.TB) (context.Context, *pgxpool.Pool) {
	tb.Helper()
	testutil.SilenceStandardLogger(tb)

	ctx := context.Background()

	cfgPath := getConfigPath(tb)
	cfg, err := config.GetConfig(cfgPath)
	require.NoError(tb, err, "failed to load config: %s", cfgPath)

	pool, err := cfg.DB.CreatePool(ctx)
	require.NoError(tb, err)
	tb.Cleanup(func() { pool.Close() })

	require.NoError(tb, pool.Ping(ctx))
	return ctx, pool
}

func doEchoQuery(ctx context.Context, pool *pgxpool.Pool, value int) error {
	var result int
	if err := pool.QueryRow(ctx, "SELECT $1::int", value).Scan(&result); err != nil {
		return err
	}
	if result != value {
		return fmt.Errorf("expected %d, got %d", value, result)
	}
	return nil
}

func runConcurrentEchoQueries(ctx context.Context, pool *pgxpool.Pool, workers int) []error {
	var wg sync.WaitGroup
	errCh := make(chan error, workers)

	for i := 0; i < workers; i++ {
		i := i
		wg.Go(func() {
			if err := doEchoQuery(ctx, pool, i); err != nil {
				errCh <- err
			}
		})
	}

	wg.Wait()
	close(errCh)

	errs := make([]error, 0, len(errCh))
	for err := range errCh {
		errs = append(errs, err)
	}

	return errs
}

// TestConcurrentEchoQueries validates correctness of parallel read-only queries.
func TestConcurrentEchoQueries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, pool := mustCreatePool(t)

	t.Run("100 workers", func(t *testing.T) {
		errs := runConcurrentEchoQueries(ctx, pool, 100)
		require.Empty(t, errs, "expected no errors, got %d", len(errs))
	})
}

// TestConcurrentEchoQueriesAcrossLoadLevels checks DB behavior under different concurrency levels.
func TestConcurrentEchoQueriesAcrossLoadLevels(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, pool := mustCreatePool(t)

	loadLevels := []int{10, 50, 100, 250}
	for _, workers := range loadLevels {
		workers := workers
		t.Run(fmt.Sprintf("%d_workers", workers), func(t *testing.T) {
			start := time.Now()
			errs := runConcurrentEchoQueries(ctx, pool, workers)
			elapsed := time.Since(start)

			require.Empty(t, errs, "expected no errors, got %d", len(errs))
			t.Logf("workers=%d elapsed=%s rps=%.2f", workers, elapsed, float64(workers)/elapsed.Seconds())
		})
	}
}

// TestConnectionAcquireLatency validates that acquiring and releasing pooled connections is stable.
func TestConnectionAcquireLatency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, pool := mustCreatePool(t)

	const attempts = 50
	var total time.Duration

	for i := 0; i < attempts; i++ {
		start := time.Now()
		conn, err := pool.Acquire(ctx)
		require.NoError(t, err)
		conn.Release()
		total += time.Since(start)
	}

	avg := total / attempts
	t.Logf("acquire attempts=%d avg_latency=%s", attempts, avg)
	require.Less(t, avg, 150*time.Millisecond, "average acquire latency is too high")
}

// TestTransactionRollbackIntegrity verifies rollback keeps DB state unchanged.
func TestTransactionRollbackIntegrity(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, pool := mustCreatePool(t)

	var before int
	require.NoError(t, pool.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&before))

	tx, err := pool.Begin(ctx)
	require.NoError(t, err)

	_, err = tx.Exec(ctx, "INSERT INTO users (username, hashed_password) VALUES ($1, $2)", "rollback_user_perf", "hash")
	require.NoError(t, err)
	require.NoError(t, tx.Rollback(ctx))

	var after int
	require.NoError(t, pool.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&after))
	require.Equal(t, before, after)
}
