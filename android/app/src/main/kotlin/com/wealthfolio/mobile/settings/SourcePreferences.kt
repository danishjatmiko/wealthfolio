package com.wealthfolio.mobile.settings

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

private val Context.sourcePrefsDataStore: DataStore<Preferences> by preferencesDataStore(name = "source_prefs")

/** Per-source (by catalog `source` id, e.g. "gopay") on/off toggle, checked
 * by TransactionNotificationListener before it reads any notification
 * content for that package. Defaults to off for every source — the user
 * opts each one in explicitly in Settings, rather than the app reading
 * everything from day one. Keyed by the source string directly, not a
 * NotificationSource enum — the catalog is fetched from the backend now,
 * so there's no fixed set of sources to enumerate at compile time. */
@Singleton
class SourcePreferences @Inject constructor(@ApplicationContext private val context: Context) {

    private fun key(source: String) = booleanPreferencesKey("enabled_$source")

    fun isEnabled(source: String): Flow<Boolean> =
        context.sourcePrefsDataStore.data.map { it[key(source)] ?: false }

    suspend fun setEnabled(source: String, enabled: Boolean) {
        context.sourcePrefsDataStore.edit { it[key(source)] = enabled }
    }
}
