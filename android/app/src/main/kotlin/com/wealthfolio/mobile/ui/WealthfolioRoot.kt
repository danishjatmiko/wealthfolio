package com.wealthfolio.mobile.ui

import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import com.wealthfolio.mobile.auth.LoginScreen
import com.wealthfolio.mobile.auth.TokenStore

@Composable
fun WealthfolioRoot(tokenStore: TokenStore) {
    val isLoggedIn by tokenStore.isLoggedIn.collectAsState()
    if (isLoggedIn) {
        MainShell(tokenStore = tokenStore)
    } else {
        LoginScreen()
    }
}
