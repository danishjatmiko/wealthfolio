package com.wealthfolio.mobile.auth

import javax.inject.Inject
import okhttp3.Interceptor
import okhttp3.Response

/**
 * Attaches the stored session token as `Cookie: wf_session=<token>` on
 * every request — the backend's AuthMiddleware just reads that header
 * regardless of whether a real cookie jar or this interceptor put it
 * there (see backend httpapi/middleware.go), so nothing server-side had
 * to change for the app to use the same session mechanism as the web
 * frontend. A 401 means the session is gone (expired past its absolute
 * cap, or revoked) — clearing the token here, in one place, is what every
 * caller (login screen redirect, sync worker's "stop this sweep" logic)
 * ends up reacting to via TokenStore.isLoggedIn.
 */
class AuthInterceptor @Inject constructor(private val tokenStore: TokenStore) : Interceptor {
    override fun intercept(chain: Interceptor.Chain): Response {
        val token = tokenStore.token
        val request = if (token != null) {
            chain.request().newBuilder()
                .addHeader("Cookie", "wf_session=$token")
                .build()
        } else {
            chain.request()
        }

        val response = chain.proceed(request)
        if (response.code == 401) {
            tokenStore.clear()
        }
        return response
    }
}
