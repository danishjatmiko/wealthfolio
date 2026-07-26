package com.wealthfolio.mobile.sync

import android.content.Context
import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.booleanPreferencesKey
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.preferencesDataStore
import dagger.hilt.android.qualifiers.ApplicationContext
import javax.inject.Inject
import javax.inject.Singleton
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map

private val Context.syncPrefsDataStore: DataStore<Preferences> by preferencesDataStore(name = "sync_prefs")

/** Whether the daily auto-clear of resolved (SENT/IGNORED) Outbox rows is
 * on — see SyncScheduler.scheduleAutoClear/cancelAutoClear and
 * OutboxAutoClearWorker. Defaults off. */
@Singleton
class SyncPreferences @Inject constructor(@ApplicationContext private val context: Context) {

    private val autoClearEnabledKey = booleanPreferencesKey("auto_clear_enabled")

    val autoClearEnabled: Flow<Boolean> =
        context.syncPrefsDataStore.data.map { it[autoClearEnabledKey] ?: false }

    suspend fun setAutoClearEnabled(enabled: Boolean) {
        context.syncPrefsDataStore.edit { it[autoClearEnabledKey] = enabled }
    }
}
