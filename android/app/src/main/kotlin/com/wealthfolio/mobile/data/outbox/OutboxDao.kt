package com.wealthfolio.mobile.data.outbox

import androidx.room.Dao
import androidx.room.Insert
import androidx.room.OnConflictStrategy
import androidx.room.Query
import androidx.room.Update
import kotlinx.coroutines.flow.Flow

@Dao
interface OutboxDao {
    /** Ignored on conflict rather than replacing — idempotencyKey is
     * UNIQUE, and if a row already exists for it we want to keep whatever
     * status it's already progressed to, not reset it back to PENDING. */
    @Insert(onConflict = OnConflictStrategy.IGNORE)
    suspend fun insert(expense: OutboxExpense): Long

    @Update
    suspend fun update(expense: OutboxExpense)

    @Query("SELECT * FROM outbox_expenses WHERE status IN ('PENDING', 'FAILED') ORDER BY createdAt")
    suspend fun listRetryable(): List<OutboxExpense>

    @Query("SELECT * FROM outbox_expenses ORDER BY createdAt DESC")
    fun observeAll(): Flow<List<OutboxExpense>>

    @Query("SELECT * FROM outbox_expenses WHERE id = :id")
    suspend fun getById(id: Long): OutboxExpense?
}
