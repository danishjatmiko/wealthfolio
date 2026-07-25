package com.wealthfolio.mobile.data.notificationcatalog

import androidx.room.Database
import androidx.room.RoomDatabase

/**
 * Its own Room database, separate from OutboxDatabase, so this catalog's
 * schema (which only ever holds a full-replace mirror of server data, no
 * locally-generated state) never bumps the version on the outbox DB, which
 * holds real synced/queued expense data.
 */
@Database(entities = [NotificationAppEntity::class, NotificationAppPackageEntity::class], version = 1, exportSchema = false)
abstract class NotificationCatalogDatabase : RoomDatabase() {
    abstract fun notificationCatalogDao(): NotificationCatalogDao
}
