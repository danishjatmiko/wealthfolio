package com.wealthfolio.mobile.sync

import androidx.work.BackoffPolicy
import androidx.work.Constraints
import androidx.work.ExistingPeriodicWorkPolicy
import androidx.work.ExistingWorkPolicy
import androidx.work.NetworkType
import androidx.work.OneTimeWorkRequestBuilder
import androidx.work.PeriodicWorkRequestBuilder
import androidx.work.WorkManager
import androidx.work.WorkRequest
import java.util.concurrent.TimeUnit
import javax.inject.Inject
import javax.inject.Singleton

/**
 * Two entry points into the same OutboxSyncWorker: a periodic ~15-minute
 * sweep (WorkManager's floor for PeriodicWorkRequest) that's the
 * belt-and-suspenders answer to "even if you close the app, it keeps
 * retrying" (req. #7), and an expedited one-shot fired right after a
 * notification is captured so it doesn't sit for up to 15 minutes before
 * its first attempt.
 */
@Singleton
class SyncScheduler @Inject constructor(private val workManager: WorkManager) {

    private val networkConstraint = Constraints.Builder()
        .setRequiredNetworkType(NetworkType.CONNECTED)
        .build()

    fun schedulePeriodicSweep() {
        val request = PeriodicWorkRequestBuilder<OutboxSyncWorker>(15, TimeUnit.MINUTES)
            .setConstraints(networkConstraint)
            .setBackoffCriteria(BackoffPolicy.EXPONENTIAL, WorkRequest.MIN_BACKOFF_MILLIS, TimeUnit.MILLISECONDS)
            .build()
        workManager.enqueueUniquePeriodicWork(
            PERIODIC_WORK_NAME,
            ExistingPeriodicWorkPolicy.KEEP,
            request,
        )
    }

    fun syncNow() {
        val request = OneTimeWorkRequestBuilder<OutboxSyncWorker>()
            .setConstraints(networkConstraint)
            .setExpedited(androidx.work.OutOfQuotaPolicy.RUN_AS_NON_EXPEDITED_WORK_REQUEST)
            .setBackoffCriteria(BackoffPolicy.EXPONENTIAL, WorkRequest.MIN_BACKOFF_MILLIS, TimeUnit.MILLISECONDS)
            .build()
        workManager.enqueueUniqueWork(ONE_SHOT_WORK_NAME, ExistingWorkPolicy.REPLACE, request)
    }

    /** Called when the Sync status screen's "Auto-clear daily" toggle
     * turns on. REPLACE rather than KEEP (unlike [schedulePeriodicSweep])
     * — toggling off then back on is a fresh explicit user action each
     * time, so it should reset the 24h timer rather than resuming
     * whatever schedule was already there. */
    fun scheduleAutoClear() {
        val request = PeriodicWorkRequestBuilder<OutboxAutoClearWorker>(1, TimeUnit.DAYS).build()
        workManager.enqueueUniquePeriodicWork(
            AUTO_CLEAR_WORK_NAME,
            ExistingPeriodicWorkPolicy.REPLACE,
            request,
        )
    }

    /** Called when the toggle turns off. */
    fun cancelAutoClear() {
        workManager.cancelUniqueWork(AUTO_CLEAR_WORK_NAME)
    }

    private companion object {
        const val PERIODIC_WORK_NAME = "outbox_periodic_sweep"
        const val ONE_SHOT_WORK_NAME = "outbox_one_shot_sync"
        const val AUTO_CLEAR_WORK_NAME = "outbox_auto_clear"
    }
}
