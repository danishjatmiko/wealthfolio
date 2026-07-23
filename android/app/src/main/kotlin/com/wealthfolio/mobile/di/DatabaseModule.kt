package com.wealthfolio.mobile.di

import android.content.Context
import androidx.room.Room
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
        Room.databaseBuilder(context, OutboxDatabase::class.java, "wealthfolio_outbox.db").build()

    @Provides
    fun provideOutboxDao(database: OutboxDatabase): OutboxDao = database.outboxDao()
}
