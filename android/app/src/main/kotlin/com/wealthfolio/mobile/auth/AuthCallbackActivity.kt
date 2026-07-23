package com.wealthfolio.mobile.auth

import android.content.Intent
import android.os.Bundle
import androidx.activity.ComponentActivity
import com.wealthfolio.mobile.MainActivity
import dagger.hilt.android.AndroidEntryPoint
import javax.inject.Inject

/**
 * No-UI activity that catches the wealthfolio://auth-callback?token=...
 * deep link the backend redirects a mobile Google login to (see
 * AuthRepository.googleLoginUrl and backend httpapi/auth.go's isMobile
 * branch), stores the token exactly like a password login would, and
 * hands off to MainActivity. Registered with android:theme
 * Theme.Wealthfolio.Transparent so it never visibly flashes on screen.
 */
@AndroidEntryPoint
class AuthCallbackActivity : ComponentActivity() {

    @Inject
    lateinit var authRepository: AuthRepository

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        val token = intent?.data?.getQueryParameter("token")
        if (token != null) {
            authRepository.completeGoogleLogin(token)
        }

        startActivity(
            Intent(this, MainActivity::class.java).addFlags(
                Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TASK,
            ),
        )
        finish()
    }
}
