package com.wealthfolio.mobile.di

import android.content.Context
import androidx.room.Room
import com.wealthfolio.mobile.data.notificationcatalog.NotificationCatalogDao
import com.wealthfolio.mobile.data.notificationcatalog.NotificationCatalogDatabase
import com.wealthfolio.mobile.data.outbox.OutboxDao
import com.wealthfolio.mobile.data.outbox.OutboxDatabase
import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.android.qualifiers.ApplicationContext
import dagger.hilt.components.SingletonComponent
import javax.inject.Singleton

@Module
@InstallIn(SingletonComponent::class)
object DatabaseModule {

    @Provides
    @Singleton
    fun provideOutboxDatabase(@ApplicationContext context: Context): OutboxDatabase =
        Room.databaseBuilder(context, OutboxDatabase::class.java, "wealthfolio_outbox.db")
            // This table is just a local sync-status log (the real budget
            // data lives server-side) — safe to wipe and rebuild from
            // future captures on a schema bump, so no Migration objects.
            .fallbackToDestructiveMigration(dropAllTables = true)
            .build()

    @Provides
    fun provideOutboxDao(database: OutboxDatabase): OutboxDao = database.outboxDao()

    @Provides
    @Singleton
    fun provideNotificationCatalogDatabase(@ApplicationContext context: Context): NotificationCatalogDatabase =
        Room.databaseBuilder(context, NotificationCatalogDatabase::class.java, "wealthfolio_notification_catalog.db").build()

    @Provides
    fun provideNotificationCatalogDao(database: NotificationCatalogDatabase): NotificationCatalogDao =
        database.notificationCatalogDao()
}
