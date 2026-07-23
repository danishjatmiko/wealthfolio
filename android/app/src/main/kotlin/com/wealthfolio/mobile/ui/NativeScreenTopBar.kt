package com.wealthfolio.mobile.ui

import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material3.CenterAlignedTopAppBar
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.runtime.Composable

/** Shared top bar for Sync Status / Settings — both are reached only via
 * the Web tab's native-link buttons (see WebTabScreen's scheme
 * interception), not a persistent tab, so each needs an explicit way back
 * rather than relying solely on the system back gesture. */
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun NativeScreenTopBar(title: String, onBack: () -> Unit) {
    CenterAlignedTopAppBar(
        title = { Text(title) },
        navigationIcon = {
            IconButton(onClick = onBack) {
                Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Back to Wealthfolio")
            }
        },
        colors = TopAppBarDefaults.centerAlignedTopAppBarColors(),
    )
}
