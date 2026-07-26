package com.wealthfolio.mobile.sync

import android.content.Context
import androidx.hilt.work.HiltWorker
import androidx.work.CoroutineWorker
import androidx.work.WorkerParameters
import com.wealthfolio.mobile.data.outbox.OutboxRepository
import dagger.assisted.Assisted
import dagger.assisted.AssistedInject

/**
 * Daily housekeeping for the Sync status screen — clears resolved
 * (SENT/IGNORED) Outbox rows only, same scope as the manual "Clear
 * resolved" button. Never touches PENDING/FAILED: those haven't reached
 * the server yet, and this runs unattended, so silently discarding one
 * would permanently lose a real captured expense with no chance to
 * notice. Scheduled/cancelled by SyncScheduler, gated on
 * SyncPreferences.autoClearEnabled.
 */
@HiltWorker
class OutboxAutoClearWorker @AssistedInject constructor(
    @Assisted context: Context,
    @Assisted params: WorkerParameters,
    private val outboxRepository: OutboxRepository,
) : CoroutineWorker(context, params) {

    override suspend fun doWork(): Result {
        outboxRepository.clearResolved()
        return Result.success()
    }
}
