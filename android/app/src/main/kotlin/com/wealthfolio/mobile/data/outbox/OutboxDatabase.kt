package com.wealthfolio.mobile.data.outbox

import androidx.room.Database
import androidx.room.RoomDatabase
import androidx.room.TypeConverters

@Database(entities = [OutboxExpense::class], version = 1, exportSchema = false)
@TypeConverters(Converters::class)
abstract class OutboxDatabase : RoomDatabase() {
    abstract fun outboxDao(): OutboxDao
}
