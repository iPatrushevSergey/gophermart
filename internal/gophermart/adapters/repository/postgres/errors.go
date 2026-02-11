package postgres

import (
	"errors"
	"io"
	"net"
	"os"
	"syscall"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// IsRetriable checks whether a PostgreSQL or network error is transient and the operation can be retried.
// Reference: https://www.postgresql.org/docs/current/errcodes-appendix.html
func IsRetriable(err error) bool {
	if err == nil {
		return false
	}

	// Network-level errors (connection refused, timeout, broken pipe, EOF)
	if isNetworkError(err) {
		return true
	}

	// PostgreSQL-level errors
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}

	code := pgErr.Code

	// Class 08 — Connection Exception (connection lost, broken pipe)
	if pgerrcode.IsConnectionException(code) {
		return true
	}

	// Class 25 — Invalid Transaction State (transaction was aborted externally)
	if pgerrcode.IsInvalidTransactionState(code) {
		return true
	}

	// Class 40 — Transaction Rollback (deadlock, serialization failure)
	if pgerrcode.IsTransactionRollback(code) {
		return true
	}

	// Class 53 — Insufficient Resources (out of memory, too many connections)
	if pgerrcode.IsInsufficientResources(code) {
		return true
	}

	// Class 57 — Operator Intervention (server shutting down, crash recovery)
	switch code {
	case pgerrcode.AdminShutdown,
		pgerrcode.CrashShutdown,
		pgerrcode.CannotConnectNow,
		pgerrcode.DatabaseDropped,
		pgerrcode.IdleInTransactionSessionTimeout:
		return true
	}

	return false
}

// isNetworkError checks for common transient network errors.
func isNetworkError(err error) bool {
	// EOF — connection was closed by the server
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}

	// Connection refused, reset, broken pipe
	if errors.Is(err, syscall.ECONNREFUSED) ||
		errors.Is(err, syscall.ECONNRESET) ||
		errors.Is(err, syscall.EPIPE) {
		return true
	}

	// OS-level timeout
	if errors.Is(err, os.ErrDeadlineExceeded) {
		return true
	}

	// net.Error with Timeout (e.g. dial timeout, read timeout)
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	return false
}
