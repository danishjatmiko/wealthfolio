package com.wealthfolio.mobile.settings

import android.content.Context
import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.booleanPreferencesKey
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.preferencesDataStore
import com.wealthfolio.mobile.notifications.NotificationSource
import dagger.hilt.android.qualifiers.ApplicationContext
import javax.inject.Inject
import javax.inject.Singleton
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map

private val Context.sourcePrefsDataStore: DataStore<Preferences> by preferencesDataStore(name = "source_prefs")

/** Per-source on/off toggle, checked by TransactionNotificationListener
 * before it reads any notification content for that package. Defaults to
 * off for every source — the user opts each one in explicitly in
 * Settings, rather than the app reading everything from day one. */
@Singleton
class SourcePreferences @Inject constructor(@ApplicationContext private val context: Context) {

    private fun key(source: NotificationSource) = booleanPreferencesKey("enabled_${source.id}")

    fun isEnabled(source: NotificationSource): Flow<Boolean> =
        context.sourcePrefsDataStore.data.map { it[key(source)] ?: false }

    fun enabledSources(): Flow<Set<NotificationSource>> =
        context.sourcePrefsDataStore.data.map { prefs ->
            NotificationSource.entries.filterTo(mutableSetOf()) { prefs[key(it)] == true }
        }

    suspend fun setEnabled(source: NotificationSource, enabled: Boolean) {
        context.sourcePrefsDataStore.edit { it[key(source)] = enabled }
    }
}
