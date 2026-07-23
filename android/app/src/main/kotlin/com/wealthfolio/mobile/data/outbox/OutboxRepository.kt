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
        dao.update(expense.copy(status = OutboxStatus.SENT, lastError = null))
    }

    suspend fun markIgnored(expense: OutboxExpense) {
        dao.update(expense.copy(status = OutboxStatus.IGNORED, lastError = null))
    }

    suspend fun markFailed(expense: OutboxExpense, error: String) {
        dao.update(
            expense.copy(
                status = OutboxStatus.FAILED,
                attemptCount = expense.attemptCount + 1,
                lastError = error,
            ),
        )
    }

    suspend fun markRetryPending(expense: OutboxExpense, error: String?) {
        dao.update(
            expense.copy(
                status = OutboxStatus.PENDING,
                attemptCount = expense.attemptCount + 1,
                lastError = error,
            ),
        )
    }

    /** Used by the Sync Status screen's per-item "Retry now" for a FAILED
     * row — puts it back in the retryable set for the next sweep. */
    suspend fun requeue(id: Long) {
        val expense = dao.getById(id) ?: return
        dao.update(expense.copy(status = OutboxStatus.PENDING))
    }
}
