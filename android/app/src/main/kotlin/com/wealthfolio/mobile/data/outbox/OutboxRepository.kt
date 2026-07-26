package com.wealthfolio.mobile.data.outbox

import javax.inject.Inject
import javax.inject.Singleton
import kotlinx.coroutines.flow.Flow

@Singleton
class OutboxRepository @Inject constructor(private val dao: OutboxDao) {
    fun observeAll(): Flow<List<OutboxExpense>> = dao.observeAll()

    suspend fun enqueue(expense: OutboxExpense): Long = dao.insert(expense)

    suspend fun listRetryable(): List<OutboxExpense> = dao.listRetryable()

    suspend fun markSent(expense: OutboxExpense) {
        dao.update(expense.copy(status = OutboxStatus.SENT, lastError = null, updatedAt = System.currentTimeMillis()))
    }

    suspend fun markIgnored(expense: OutboxExpense) {
        dao.update(expense.copy(status = OutboxStatus.IGNORED, lastError = null, updatedAt = System.currentTimeMillis()))
    }

    suspend fun markFailed(expense: OutboxExpense, error: String) {
        dao.update(
            expense.copy(
                status = OutboxStatus.FAILED,
                attemptCount = expense.attemptCount + 1,
                lastError = error,
                updatedAt = System.currentTimeMillis(),
            ),
        )
    }

    suspend fun markRetryPending(expense: OutboxExpense, error: String?) {
        dao.update(
            expense.copy(
                status = OutboxStatus.PENDING,
                attemptCount = expense.attemptCount + 1,
                lastError = error,
                updatedAt = System.currentTimeMillis(),
            ),
        )
    }

    /** Used by the Sync Status screen's per-item "Retry now" for a FAILED
     * row — puts it back in the retryable set for the next sweep. */
    suspend fun requeue(id: Long) {
        val expense = dao.getById(id) ?: return
        dao.update(expense.copy(status = OutboxStatus.PENDING, updatedAt = System.currentTimeMillis()))
    }

    /** Manual "Clear resolved" button, and the daily auto-clear worker —
     * only ever removes SENT/IGNORED rows, never PENDING/FAILED. */
    suspend fun clearResolved() = dao.deleteResolved()

    /** Manual "Clear all" button only — removes every row regardless of
     * status, including anything not yet synced. */
    suspend fun clearAll() = dao.deleteAll()
}
