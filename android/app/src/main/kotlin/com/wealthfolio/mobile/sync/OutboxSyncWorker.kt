package com.wealthfolio.mobile.sync

import android.content.Context
import androidx.hilt.work.HiltWorker
import androidx.work.CoroutineWorker
import androidx.work.WorkerParameters
import com.wealthfolio.mobile.data.outbox.OutboxExpense
import com.wealthfolio.mobile.data.outbox.OutboxRepository
import com.wealthfolio.mobile.network.ApiService
import com.wealthfolio.mobile.network.dto.IngestExpenseRequest
import dagger.assisted.Assisted
import dagger.assisted.AssistedInject
import java.net.HttpURLConnection

/**
 * Drains the outbox: POSTs every PENDING/FAILED row to /expense-ingestions
 * and applies the status transitions from the plan (Part B3) —
 *   200 "created" -> SENT
 *   200 "ignored" -> IGNORED (terminal, not retried again)
 *   422           -> FAILED (retryable — resolves once the user fixes
 *                    their period/mapping/envelope setup, so the next
 *                    sweep tries it again)
 *   network / 5xx -> stays PENDING, attemptCount++
 *   401           -> stop this sweep entirely; AuthInterceptor already
 *                    cleared the token, so TokenStore.isLoggedIn flips
 *                    and the UI routes back to login on its own.
 * Runs both as a periodic sweep and as an expedited one-shot fired right
 * after a notification is captured — see SyncScheduler.
 */
@HiltWorker
class OutboxSyncWorker @AssistedInject constructor(
    @Assisted context: Context,
    @Assisted params: WorkerParameters,
    private val outboxRepository: OutboxRepository,
    private val api: ApiService,
) : CoroutineWorker(context, params) {

    override suspend fun doWork(): Result {
        val pending = outboxRepository.listRetryable()
        for (expense in pending) {
            val stop = syncOne(expense)
            if (stop) return Result.success()
        }
        return Result.success()
    }

    /** Returns true if the whole sweep should stop (session expired). */
    private suspend fun syncOne(expense: OutboxExpense): Boolean {
        val response = try {
            api.ingestExpense(
                IngestExpenseRequest(
                    idempotencyKey = expense.idempotencyKey,
                    source = expense.source,
                    rawTitle = expense.rawTitle,
                    rawText = expense.rawText,
                    rawBigText = expense.rawBigText,
                    occurredAt = expense.occurredAt,
                ),
            )
        } catch (e: Exception) {
            outboxRepository.markRetryPending(expense, e.message ?: "network error")
            return false
        }

        if (response.code() == HttpURLConnection.HTTP_UNAUTHORIZED) {
            return true
        }

        val body = response.body()
        if (response.isSuccessful && body != null) {
            when (body.status) {
                "created" -> outboxRepository.markSent(expense)
                "ignored" -> outboxRepository.markIgnored(expense)
                else -> outboxRepository.markRetryPending(expense, "unexpected status: ${body.status}")
            }
            return false
        }

        if (response.code() == 422) {
            outboxRepository.markFailed(expense, response.errorBody()?.string() ?: "unprocessable")
        } else {
            outboxRepository.markRetryPending(expense, "server error (${response.code()})")
        }
        return false
    }
}
