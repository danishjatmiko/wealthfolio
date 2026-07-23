package com.wealthfolio.mobile.auth

import android.content.Context
import android.content.SharedPreferences
import androidx.security.crypto.EncryptedSharedPreferences
import androidx.security.crypto.MasterKey
import dagger.hilt.android.qualifiers.ApplicationContext
import javax.inject.Inject
import javax.inject.Singleton
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

/**
 * Keystore-backed storage for the session token — the same opaque string
 * the backend's `sessions.token` uses, whether it came from password login
 * or the Google mobile deep-link callback (see AuthRepository). Read by
 * AuthInterceptor on every request and by WebTabScreen to seed the
 * WebView's cookie jar; nothing else in the app touches raw prefs.
 */
@Singleton
class TokenStore @Inject constructor(@ApplicationContext context: Context) {

    private val prefs: SharedPreferences by lazy {
        val masterKey = MasterKey.Builder(context)
            .setKeyScheme(MasterKey.KeyScheme.AES256_GCM)
            .build()
        EncryptedSharedPreferences.create(
            context,
            "wealthfolio_secure_prefs",
            masterKey,
            EncryptedSharedPreferences.PrefKeyEncryptionScheme.AES256_SIV,
            EncryptedSharedPreferences.PrefValueEncryptionScheme.AES256_GCM,
        )
    }

    private val _isLoggedIn = MutableStateFlow(false)

    /** Reactive login state — MainActivity's nav graph and the "session
     * expired" banner both observe this rather than polling [token]. */
    val isLoggedIn: StateFlow<Boolean> = _isLoggedIn.asStateFlow()

    var token: String?
        get() = prefs.getString(KEY_TOKEN, null)
        set(value) {
            prefs.edit().putString(KEY_TOKEN, value).apply()
            _isLoggedIn.value = !value.isNullOrEmpty()
        }

    init {
        _isLoggedIn.value = !prefs.getString(KEY_TOKEN, null).isNullOrEmpty()
    }

    fun clear() {
        token = null
    }

    private companion object {
        const val KEY_TOKEN = "session_token"
    }
}
