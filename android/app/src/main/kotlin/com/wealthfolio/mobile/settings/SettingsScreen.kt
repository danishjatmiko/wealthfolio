package com.wealthfolio.mobile.settings

import android.content.Intent
import android.provider.Settings
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Card
import androidx.compose.material3.DropdownMenu
import androidx.compose.material3.DropdownMenuItem
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.ExposedDropdownMenuBox
import androidx.compose.material3.ExposedDropdownMenuDefaults
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Switch
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.LocalLifecycleOwner
import androidx.compose.ui.unit.dp
import androidx.core.app.NotificationManagerCompat
import androidx.compose.material3.Scaffold
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.LifecycleEventObserver
import com.wealthfolio.mobile.notifications.NotificationSource
import com.wealthfolio.mobile.ui.NativeScreenTopBar

/** Whether the OS's "notification access" permission (system-wide, not
 * per-source — see TransactionNotificationListener) is currently granted
 * to this app. */
private fun isNotificationAccessGranted(context: android.content.Context): Boolean =
    NotificationManagerCompat.getEnabledListenerPackages(context).contains(context.packageName)

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun SettingsScreen(onBack: () -> Unit, viewModel: SettingsViewModel = hiltViewModel()) {
    val state by viewModel.uiState.collectAsState()
    val context = LocalContext.current

    var notificationAccessGranted by remember { mutableStateOf(isNotificationAccessGranted(context)) }
    val lifecycleOwner = LocalLifecycleOwner.current
    DisposableEffect(lifecycleOwner) {
        // Re-check on every resume, not just once — the only way to grant
        // this is a round trip through the system settings screen, so this
        // is what notices you came back with it turned on (or off).
        val observer = LifecycleEventObserver { _, event ->
            if (event == Lifecycle.Event.ON_RESUME) {
                notificationAccessGranted = isNotificationAccessGranted(context)
            }
        }
        lifecycleOwner.lifecycle.addObserver(observer)
        onDispose { lifecycleOwner.lifecycle.removeObserver(observer) }
    }

    Scaffold(topBar = { NativeScreenTopBar("Settings", onBack) }) { padding ->
        Column(modifier = Modifier.fillMaxSize().padding(padding).padding(16.dp)) {
            Text("Notification sources", style = MaterialTheme.typography.titleLarge)
            Text(
                "Turn a source on, then pick which envelope its captured expenses auto-file into.",
                style = MaterialTheme.typography.bodyMedium,
            )
            Spacer(Modifier.height(8.dp))
            if (notificationAccessGranted) {
                Text(
                    "✓ Notification access granted",
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.primary,
                )
            } else {
                OutlinedButton(onClick = {
                    context.startActivity(Intent(Settings.ACTION_NOTIFICATION_LISTENER_SETTINGS))
                }) {
                    Text("Grant notification access")
                }
            }
            Spacer(Modifier.height(16.dp))

            if (state.error != null) {
                Text(state.error!!, color = MaterialTheme.colorScheme.error)
                Spacer(Modifier.height(8.dp))
            }

            LazyColumn(modifier = Modifier.weight(1f)) {
                items(state.rows) { row ->
                    SourceCard(
                        row = row,
                        availableEnvelopeNames = state.availableEnvelopeNames,
                        onEnabledChange = { viewModel.setSourceEnabled(row.source, it) },
                        onEnvelopeSelected = { viewModel.setEnvelopeMapping(row.source, it) },
                    )
                    Spacer(Modifier.height(12.dp))
                }
            }

            HorizontalDivider(modifier = Modifier.padding(vertical = 8.dp))
            OutlinedButton(
                onClick = { viewModel.logout() },
                colors = ButtonDefaults.outlinedButtonColors(contentColor = MaterialTheme.colorScheme.error),
                modifier = Modifier.fillMaxWidth(),
            ) {
                Text("Sign out")
            }
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun SourceCard(
    row: SourceRow,
    availableEnvelopeNames: List<String>,
    onEnabledChange: (Boolean) -> Unit,
    onEnvelopeSelected: (String) -> Unit,
) {
    var expanded by remember { mutableStateOf(false) }

    Card(modifier = Modifier.fillMaxWidth().padding(4.dp)) {
        Column(modifier = Modifier.padding(16.dp)) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
            ) {
                Text(row.source.displayName(), style = MaterialTheme.typography.titleMedium)
                Switch(checked = row.enabled, onCheckedChange = onEnabledChange)
            }

            Spacer(Modifier.height(8.dp))

            ExposedDropdownMenuBox(expanded = expanded, onExpandedChange = { expanded = it }) {
                OutlinedTextField(
                    value = row.mappedEnvelopeName ?: "Choose an envelope",
                    onValueChange = {},
                    readOnly = true,
                    label = { Text("Envelope") },
                    trailingIcon = { ExposedDropdownMenuDefaults.TrailingIcon(expanded = expanded) },
                    modifier = Modifier.fillMaxWidth().menuAnchor(),
                )
                DropdownMenu(
                    expanded = expanded,
                    onDismissRequest = { expanded = false },
                ) {
                    availableEnvelopeNames.forEach { name ->
                        DropdownMenuItem(
                            text = { Text(name) },
                            onClick = {
                                onEnvelopeSelected(name)
                                expanded = false
                            },
                        )
                    }
                }
            }
        }
    }
}

private fun NotificationSource.displayName(): String = when (this) {
    NotificationSource.GOPAY -> "GoPay"
    NotificationSource.DANA -> "DANA"
    NotificationSource.BCA -> "BCA / m-BCA"
}
