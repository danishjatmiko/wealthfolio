package com.wealthfolio.mobile.notifications

import android.app.Notification
import android.service.notification.NotificationListenerService
import android.service.notification.StatusBarNotification
import com.wealthfolio.mobile.data.outbox.OutboxExpense
import com.wealthfolio.mobile.data.outbox.OutboxRepository
import com.wealthfolio.mobile.settings.SourcePreferences
import com.wealthfolio.mobile.sync.SyncScheduler
import dagger.hilt.android.AndroidEntryPoint
import java.text.SimpleDateFormat
import java.util.Date
import java.util.Locale
import java.util.TimeZone
import javax.inject.Inject
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.launch

/**
 * Reads notifications only from packages belonging to a source the user
 * has toggled on in Settings (see SourcePreferences) and forwards their
 * raw title/text/bigText to the outbox untouched — no parsing happens
 * on-device; that's the backend's job (see notificationparse and the
 * plan's rationale for why). Everything else — every notification from
 * every other app on the phone — is ignored in the very first check of
 * onNotificationPosted, before any of its content is read.
 *
 * The OS's "notification access" permission this service requires is
 * granted once, system-wide, covering every app's notifications; the
 * per-source toggle here is what actually limits what we act on.
 */
@AndroidEntryPoint
class TransactionNotificationListener : NotificationListenerService() {

    @Inject lateinit var sourcePreferences: SourcePreferences
    @Inject lateinit var outboxRepository: OutboxRepository
    @Inject lateinit var syncScheduler: SyncScheduler

    private val serviceScope = CoroutineScope(SupervisorJob() + Dispatchers.IO)

    override fun onNotificationPosted(sbn: StatusBarNotification) {
        val source = NotificationSource.forPackageName(sbn.packageName) ?: return

        serviceScope.launch {
            // Read fresh on every notification rather than cached across
            // the service's lifetime, so toggling a source off in
            // Settings takes effect on the very next notification.
            val enabled = sourcePreferences.isEnabled(source).first()
            if (!enabled) return@launch

            val extras = sbn.notification.extras
            val title = extras.getCharSequence(Notification.EXTRA_TITLE)?.toString()
            val text = extras.getCharSequence(Notification.EXTRA_TEXT)?.toString()
            val bigText = extras.getCharSequence(Notification.EXTRA_BIG_TEXT)?.toString()

            val idempotencyKey = buildIdempotencyKey(source.id, sbn.packageName, title, text, sbn.postTime)

            outboxRepository.enqueue(
                OutboxExpense(
                    idempotencyKey = idempotencyKey,
                    source = source.id,
                    rawTitle = title,
                    rawText = text,
                    rawBigText = bigText,
                    occurredAt = isoFormat(sbn.postTime),
                ),
            )
            syncScheduler.syncNow()
        }
    }

    private fun isoFormat(millis: Long): String {
        val format = SimpleDateFormat("yyyy-MM-dd'T'HH:mm:ss'Z'", Locale.US)
        format.timeZone = TimeZone.getTimeZone("UTC")
        return format.format(Date(millis))
    }
}
