package com.wealthfolio.mobile.settings

import android.content.Intent
import android.provider.Settings
import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.Logout
import androidx.compose.material.icons.filled.CheckCircle
import androidx.compose.material.icons.filled.Notifications
import androidx.compose.material.icons.filled.WarningAmber
import androidx.compose.material3.AssistChip
import androidx.compose.material3.AssistChipDefaults
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.DropdownMenu
import androidx.compose.material3.DropdownMenuItem
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Switch
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.LocalLifecycleOwner
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.core.app.NotificationManagerCompat
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.LifecycleEventObserver
import com.wealthfolio.mobile.notifications.NotificationSource
import com.wealthfolio.mobile.ui.NativeScreenTopBar
import com.wealthfolio.mobile.ui.theme.EthernaForest
import com.wealthfolio.mobile.ui.theme.EthernaGoldSoft
import com.wealthfolio.mobile.ui.theme.EthernaGoldText
import com.wealthfolio.mobile.ui.theme.EthernaGreenSoft
import com.wealthfolio.mobile.ui.theme.EthernaGreenText
import com.wealthfolio.mobile.ui.theme.EthernaRedSoft

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

    val hasAnyEnvelopes = state.availableEnvelopeNames.isNotEmpty()

    Scaffold(topBar = { NativeScreenTopBar("Settings", onBack) }) { padding ->
        Column(modifier = Modifier.fillMaxSize().padding(padding).padding(16.dp)) {

            LazyColumn(
                modifier = Modifier.weight(1f),
                verticalArrangement = Arrangement.spacedBy(18.dp),
            ) {
                item { AccountCard(displayName = state.displayName, email = state.email) }

                item {
                    PermissionCard(
                        granted = notificationAccessGranted,
                        onOpenSystemSettings = {
                            context.startActivity(Intent(Settings.ACTION_NOTIFICATION_LISTENER_SETTINGS))
                        },
                    )
                }

                if (state.error != null) {
                    item { Text(state.error!!, color = MaterialTheme.colorScheme.error) }
                }

                item {
                    Column {
                        Text(
                            "Notification sources",
                            style = MaterialTheme.typography.titleMedium,
                            fontWeight = FontWeight.Bold,
                        )
                        Spacer(Modifier.height(2.dp))
                        Text(
                            "Turn a source on, then choose which envelope its captured expenses auto-file into.",
                            style = MaterialTheme.typography.bodySmall,
                            color = MaterialTheme.colorScheme.onSurfaceVariant,
                        )
                    }
                }

                items(state.rows) { row ->
                    SourceCard(
                        row = row,
                        hasAnyEnvelopes = hasAnyEnvelopes,
                        availableEnvelopeNames = state.availableEnvelopeNames,
                        onEnabledChange = { viewModel.setSourceEnabled(row.source, it) },
                        onEnvelopeSelected = { viewModel.setEnvelopeMapping(row.source, it) },
                    )
                }

                item {
                    Text(
                        "Only GoPay, DANA and BCA notifications are ever read — everything else on your " +
                            "phone is ignored.",
                        style = MaterialTheme.typography.labelSmall,
                        color = MaterialTheme.colorScheme.outline,
                    )
                }
            }

            HorizontalDivider(modifier = Modifier.padding(vertical = 12.dp))
            OutlinedButton(
                onClick = { viewModel.logout() },
                colors = ButtonDefaults.outlinedButtonColors(contentColor = MaterialTheme.colorScheme.error),
                border = BorderStroke(1.dp, EthernaRedSoft),
                modifier = Modifier.fillMaxWidth(),
            ) {
                Icon(Icons.AutoMirrored.Filled.Logout, contentDescription = null, modifier = Modifier.size(16.dp))
                Spacer(Modifier.width(8.dp))
                Text("Sign out")
            }
        }
    }
}

@Composable
private fun InitialsBadge(letter: String, background: Color, modifier: Modifier = Modifier) {
    Box(
        modifier = modifier.clip(RoundedCornerShape(50)).background(background),
        contentAlignment = Alignment.Center,
    ) {
        Text(letter, color = Color.White, fontWeight = FontWeight.Bold, fontSize = 13.sp)
    }
}

@Composable
private fun AccountCard(displayName: String?, email: String?) {
    Card(
        shape = RoundedCornerShape(18.dp),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
        border = BorderStroke(1.dp, MaterialTheme.colorScheme.outlineVariant),
        modifier = Modifier.fillMaxWidth(),
    ) {
        Row(
            modifier = Modifier.padding(14.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            InitialsBadge(
                letter = (displayName?.firstOrNull() ?: '·').uppercaseChar().toString(),
                background = EthernaForest,
                modifier = Modifier.size(38.dp),
            )
            Spacer(Modifier.width(11.dp))
            Column(modifier = Modifier.weight(1f)) {
                Text(
                    displayName ?: "…",
                    style = MaterialTheme.typography.titleSmall,
                    fontWeight = FontWeight.Bold,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                )
                Text(
                    email ?: "",
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                )
            }
        }
    }
}

@Composable
private fun PermissionCard(granted: Boolean, onOpenSystemSettings: () -> Unit) {
    if (granted) {
        Card(
            shape = RoundedCornerShape(18.dp),
            colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
            border = BorderStroke(1.dp, MaterialTheme.colorScheme.outlineVariant),
            modifier = Modifier.fillMaxWidth(),
        ) {
            Column(modifier = Modifier.padding(16.dp)) {
                Row {
                    Box(
                        modifier = Modifier.size(34.dp).clip(RoundedCornerShape(10.dp)).background(EthernaGreenSoft),
                        contentAlignment = Alignment.Center,
                    ) {
                        Icon(
                            Icons.Filled.CheckCircle,
                            contentDescription = null,
                            tint = EthernaGreenText,
                            modifier = Modifier.size(18.dp),
                        )
                    }
                    Spacer(Modifier.width(10.dp))
                    Column {
                        Text(
                            "Notification access granted",
                            style = MaterialTheme.typography.titleSmall,
                            fontWeight = FontWeight.Bold,
                        )
                        Text(
                            "Etherna can read notifications from the sources you enable below.",
                            style = MaterialTheme.typography.bodySmall,
                            color = MaterialTheme.colorScheme.onSurfaceVariant,
                        )
                    }
                }
                Spacer(Modifier.height(10.dp))
                OutlinedButton(onClick = onOpenSystemSettings, modifier = Modifier.padding(start = 44.dp)) {
                    Text("Manage in system settings", style = MaterialTheme.typography.labelMedium)
                }
            }
        }
    } else {
        Card(
            shape = RoundedCornerShape(18.dp),
            colors = CardDefaults.cardColors(containerColor = EthernaForest),
            modifier = Modifier.fillMaxWidth(),
        ) {
            Column(modifier = Modifier.padding(16.dp)) {
                Row {
                    Box(
                        modifier = Modifier.size(34.dp)
                            .clip(RoundedCornerShape(10.dp))
                            .background(Color.White.copy(alpha = 0.14f)),
                        contentAlignment = Alignment.Center,
                    ) {
                        Icon(
                            Icons.Filled.Notifications,
                            contentDescription = null,
                            tint = Color.White,
                            modifier = Modifier.size(18.dp),
                        )
                    }
                    Spacer(Modifier.width(10.dp))
                    Column {
                        Text(
                            "Turn on notification access",
                            style = MaterialTheme.typography.titleSmall,
                            fontWeight = FontWeight.Bold,
                            color = Color.White,
                        )
                        Text(
                            "So Etherna can auto-capture expenses from GoPay, DANA and BCA. Only those " +
                                "three apps are read — nothing else, ever.",
                            style = MaterialTheme.typography.bodySmall,
                            color = Color.White.copy(alpha = 0.78f),
                        )
                    }
                }
                Spacer(Modifier.height(10.dp))
                Button(
                    onClick = onOpenSystemSettings,
                    colors = ButtonDefaults.buttonColors(containerColor = Color.White, contentColor = EthernaForest),
                    modifier = Modifier.padding(start = 44.dp),
                ) {
                    Text("Grant access", style = MaterialTheme.typography.labelMedium, fontWeight = FontWeight.Bold)
                }
            }
        }
    }
}

@Composable
private fun NoteRow(text: String, background: Color, content: Color) {
    Row(
        modifier = Modifier.fillMaxWidth().clip(RoundedCornerShape(11.dp)).background(background).padding(10.dp),
    ) {
        Icon(Icons.Filled.WarningAmber, contentDescription = null, tint = content, modifier = Modifier.size(14.dp))
        Spacer(Modifier.width(7.dp))
        Text(text, style = MaterialTheme.typography.labelSmall, color = content)
    }
}

@Composable
private fun SourceCard(
    row: SourceRow,
    hasAnyEnvelopes: Boolean,
    availableEnvelopeNames: List<String>,
    onEnabledChange: (Boolean) -> Unit,
    onEnvelopeSelected: (String) -> Unit,
) {
    var expanded by remember { mutableStateOf(false) }
    val blockedByNoEnvelopes = !row.enabled && !hasAnyEnvelopes

    Card(
        shape = RoundedCornerShape(18.dp),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
        border = BorderStroke(1.dp, MaterialTheme.colorScheme.outlineVariant),
        modifier = Modifier.fillMaxWidth().alpha(if (row.enabled) 1f else 0.6f),
    ) {
        Column(modifier = Modifier.padding(14.dp)) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                InitialsBadge(
                    letter = row.source.displayName().first().toString(),
                    background = row.source.chipColor(),
                    modifier = Modifier.size(32.dp),
                )
                Spacer(Modifier.width(11.dp))
                Column(modifier = Modifier.weight(1f)) {
                    Text(row.source.displayName(), style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.Bold)
                    Text(
                        row.source.category(),
                        style = MaterialTheme.typography.labelSmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                    )
                }
                Switch(
                    checked = row.enabled,
                    onCheckedChange = onEnabledChange,
                    enabled = row.enabled || hasAnyEnvelopes,
                )
            }

            if (row.enabled) {
                HorizontalDivider(modifier = Modifier.padding(vertical = 10.dp))
                val hasEnvelope = row.mappedEnvelopeName != null
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    Text(
                        "FILES INTO",
                        style = MaterialTheme.typography.labelSmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                        fontWeight = FontWeight.Bold,
                    )
                    Box {
                        AssistChip(
                            onClick = { expanded = true },
                            label = { Text(row.mappedEnvelopeName ?: "Choose an envelope") },
                            colors = AssistChipDefaults.assistChipColors(
                                containerColor = if (hasEnvelope) MaterialTheme.colorScheme.surfaceVariant else EthernaGoldSoft,
                                labelColor = if (hasEnvelope) MaterialTheme.colorScheme.onSurface else EthernaGoldText,
                            ),
                            border = null,
                        )
                        DropdownMenu(expanded = expanded, onDismissRequest = { expanded = false }) {
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
                if (!hasEnvelope) {
                    Spacer(Modifier.height(8.dp))
                    NoteRow(
                        text = "On, but not filing anywhere yet — pick an envelope above.",
                        background = EthernaGoldSoft,
                        content = EthernaGoldText,
                    )
                }
            } else if (blockedByNoEnvelopes) {
                Spacer(Modifier.height(10.dp))
                NoteRow(
                    text = "You don't have any envelopes yet, so this source can't be turned on. " +
                        "Create one in Expenses first.",
                    background = EthernaRedSoft,
                    content = MaterialTheme.colorScheme.error,
                )
            }
        }
    }
}

private fun NotificationSource.displayName(): String = when (this) {
    NotificationSource.GOPAY -> "GoPay"
    NotificationSource.DANA -> "DANA"
    NotificationSource.BCA -> "BCA / m-BCA"
}

private fun NotificationSource.category(): String = when (this) {
    NotificationSource.GOPAY, NotificationSource.DANA -> "E-wallet transactions"
    NotificationSource.BCA -> "Bank transactions"
}

private fun NotificationSource.chipColor(): Color = when (this) {
    NotificationSource.GOPAY -> Color(0xFF5B7FA6)
    NotificationSource.DANA -> Color(0xFF8A76B0)
    NotificationSource.BCA -> Color(0xFFB3402F)
}
