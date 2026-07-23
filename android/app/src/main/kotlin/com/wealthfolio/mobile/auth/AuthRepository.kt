package com.wealthfolio.mobile.auth

import com.wealthfolio.mobile.AppConfig
import com.wealthfolio.mobile.network.ApiService
import com.wealthfolio.mobile.network.dto.LoginRequest
import java.io.IOException
import javax.inject.Inject
import javax.inject.Singleton

/** Couldn't reach the backend at all, or it responded with something
 * other than 200/401 — as opposed to a clean "wrong credentials" 401.
 * Wraps the underlying cause so the UI can show it verbatim, since it's
 * usually the fastest way to spot a bad AppConfig.API_BASE_URL, a backend
 * that isn't running, or (as with plain-HTTP dev backends) cleartext
 * traffic being blocked. */
class NetworkException(cause: Throwable) : Exception(cause.message, cause)

@Singleton
class AuthRepository @Inject constructor(
    private val api: ApiService,
    private val tokenStore: TokenStore,
) {
    val isLoggedIn = tokenStore.isLoggedIn

    /** URL to open in a Chrome Custom Tab to start Google sign-in. The
     * backend branches on ?platform=android to hand the token back via
     * the wealthfolio://auth-callback deep link instead of a web cookie
     * redirect — see backend httpapi/auth.go's googleLogin/googleCallback. */
    val googleLoginUrl: String
        get() = AppConfig.API_BASE_URL.removeSuffix("api/v1/") + "api/v1/auth/google/login?platform=android"

    /** Distinguishes "the server said no" from "we couldn't even talk to
     * the server" — LoginScreen shows each with a different message, so a
     * network/config problem (wrong AppConfig.API_BASE_URL, backend down,
     * cleartext blocked, etc.) doesn't get misreported as a wrong
     * password. */
    suspend fun passwordLogin(email: String, password: String): Result<Unit> {
        val response = try {
            api.login(LoginRequest(email, password))
        } catch (e: IOException) {
            return Result.failure(NetworkException(e))
        }

        val body = response.body()
        return if (response.isSuccessful && body != null) {
            tokenStore.token = body.token
            Result.success(Unit)
        } else if (response.code() == 401) {
            Result.failure(Exception("Invalid email or password"))
        } else {
            Result.failure(NetworkException(Exception("Server returned ${response.code()}")))
        }
    }

    /** Called by AuthCallbackActivity once it's extracted the token from
     * the wealthfolio://auth-callback deep link. */
    fun completeGoogleLogin(token: String) {
        tokenStore.token = token
    }

    suspend fun logout() {
        try {
            api.logout()
        } catch (_: Exception) {
            // Best-effort — the local token is cleared either way, which is
            // what actually gates the app's own UI; a failed server-side
            // revoke just means the token dies on its own at its natural
            // expiry instead of immediately.
        }
        tokenStore.clear()
    }
}
