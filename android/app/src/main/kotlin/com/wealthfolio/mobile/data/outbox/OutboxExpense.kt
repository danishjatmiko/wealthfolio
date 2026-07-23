package com.wealthfolio.mobile.data.outbox

import androidx.room.Entity
import androidx.room.Index
import androidx.room.PrimaryKey

enum class OutboxStatus {
    PENDING,
    FAILED,
    IGNORED,
    SENT,
}

/**
 * One captured notification, persisted immediately on capture so it
 * survives app close/process death — the whole point of req. #6/#7.
 * idempotencyKey is computed once at capture time (see IdempotencyKey.kt)
 * and never recomputed, so every retry of this row sends the exact same
 * key to the backend regardless of how many attempts it takes.
 */
@Entity(
    tableName = "outbox_expenses",
    indices = [Index(value = ["idempotencyKey"], unique = true)],
)
data class OutboxExpense(
    @PrimaryKey(autoGenerate = true) val id: Long = 0,
    val idempotencyKey: String,
    val source: String,
    val rawTitle: String?,
    val rawText: String?,
    val rawBigText: String?,
    val occurredAt: String,
    val status: OutboxStatus = OutboxStatus.PENDING,
    val attemptCount: Int = 0,
    val lastError: String? = null,
    val createdAt: Long = System.currentTimeMillis(),
)
