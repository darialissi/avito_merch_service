package db

import (
	"github.com/stretchr/testify/require"
	"sync/atomic"
	"testing"
)

func BenchmarkEchoQuerySequential(b *testing.B) {
	ctx, pool := mustCreatePool(b)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		require.NoError(b, doEchoQuery(ctx, pool, i))
	}
}

func BenchmarkEchoQueryParallel(b *testing.B) {
	ctx, pool := mustCreatePool(b)

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var i int64
		for pb.Next() {
			n := int(atomic.AddInt64(&i, 1))
			require.NoError(b, doEchoQuery(ctx, pool, n))
		}
	})
}

func BenchmarkConnectionAcquireRelease(b *testing.B) {
	ctx, pool := mustCreatePool(b)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conn, err := pool.Acquire(ctx)
		require.NoError(b, err)
		conn.Release()
	}
}

func BenchmarkTransactionRoundTrip(b *testing.B) {
	ctx, pool := mustCreatePool(b)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx, err := pool.Begin(ctx)
		require.NoError(b, err)

		require.NoError(b, tx.QueryRow(ctx, "SELECT 1").Scan(new(int)))
		require.NoError(b, tx.Rollback(ctx))
	}
}
