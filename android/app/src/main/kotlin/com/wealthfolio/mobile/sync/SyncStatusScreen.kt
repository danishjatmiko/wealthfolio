package com.wealthfolio.mobile.sync

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyListScope
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.Card
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Switch
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.pulltorefresh.PullToRefreshBox
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.wealthfolio.mobile.data.outbox.OutboxExpense
import com.wealthfolio.mobile.ui.NativeScreenTopBar
import java.time.Instant
import java.time.ZoneId
import java.time.format.DateTimeFormatter
import java.time.format.FormatStyle

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun SyncStatusScreen(onBack: () -> Unit, viewModel: SyncStatusViewModel = hiltViewModel()) {
    val state by viewModel.uiState.collectAsState()
    val refreshing by viewModel.isRefreshing.collectAsState()
    var showClearResolvedConfirm by remember { mutableStateOf(false) }
    var showClearAllConfirm by remember { mutableStateOf(false) }

    Scaffold(topBar = { NativeScreenTopBar("Sync status", onBack) }) { padding ->
        Column(modifier = Modifier.fillMaxSize().padding(padding).padding(16.dp)) {
            Text("Outbox", style = MaterialTheme.typography.titleLarge)
            Spacer(Modifier.height(10.dp))

            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(8.dp),
            ) {
                OutlinedButton(onClick = { showClearResolvedConfirm = true }) {
                    Text("Clear resolved")
                }
                OutlinedButton(onClick = { showClearAllConfirm = true }) {
                    Text("Clear all")
                }
            }
            Spacer(Modifier.height(12.dp))

            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
            ) {
                Column(modifier = Modifier.weight(1f)) {
                    Text("Auto-clear daily", style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.Bold)
                    Text(
                        "Clears synced/ignored history once a day. Never touches pending or failed items.",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                    )
                }
                Switch(checked = state.autoClearEnabled, onCheckedChange = viewModel::setAutoClearEnabled)
            }
            Spacer(Modifier.height(12.dp))

            PullToRefreshBox(
                isRefreshing = refreshing,
                onRefresh = viewModel::refresh,
                modifier = Modifier.weight(1f).fillMaxWidth(),
            ) {
                LazyColumn(modifier = Modifier.fillMaxSize()) {
                    section("Pending", state.pending) { null }
                    section("Failed", state.failed) { expense -> { viewModel.retry(expense) } }
                    section("Ignored (not a recognized transaction)", state.ignored) { null }
                    section("Recently synced", state.sent) { null }
                }
            }
        }
    }

    if (showClearResolvedConfirm) {
        AlertDialog(
            onDismissRequest = { showClearResolvedConfirm = false },
            title = { Text("Clear resolved history?") },
            text = { Text("Removes synced and ignored entries from this list. Pending and failed items are left untouched.") },
            confirmButton = {
                TextButton(onClick = {
                    viewModel.clearResolved()
                    showClearResolvedConfirm = false
                }) { Text("Clear resolved") }
            },
            dismissButton = {
                TextButton(onClick = { showClearResolvedConfirm = false }) { Text("Cancel") }
            },
        )
    }

    if (showClearAllConfirm) {
        AlertDialog(
            onDismissRequest = { showClearAllConfirm = false },
            title = { Text("Clear everything?") },
            text = {
                Text(
                    "Removes every entry, including anything still pending or failed. Those haven't reached the " +
                        "server yet — clearing them here means that expense is gone for good, it will never sync.",
                )
            },
            confirmButton = {
                TextButton(onClick = {
                    viewModel.clearAll()
                    showClearAllConfirm = false
                }) { Text("Clear all") }
            },
            dismissButton = {
                TextButton(onClick = { showClearAllConfirm = false }) { Text("Cancel") }
            },
        )
    }
}

private fun LazyListScope.section(
    title: String,
    items: List<OutboxExpense>,
    retryAction: (OutboxExpense) -> (() -> Unit)?,
) {
    if (items.isEmpty()) return
    item {
        Text(title, style = MaterialTheme.typography.titleMedium)
        Spacer(Modifier.height(4.dp))
    }
    items(items) { expense ->
        OutboxRow(expense = expense, onRetry = retryAction(expense))
        Spacer(Modifier.height(8.dp))
    }
}

@Composable
private fun OutboxRow(expense: OutboxExpense, onRetry: (() -> Unit)?) {
    var showDetail by remember { mutableStateOf(false) }

    Card(modifier = Modifier.fillMaxWidth().clickable { showDetail = true }) {
        Column(modifier = Modifier.padding(12.dp)) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
            ) {
                Text(expense.source.uppercase(), style = MaterialTheme.typography.labelMedium)
                if (onRetry != null) {
                    TextButton(onClick = onRetry) { Text("Retry now") }
                }
            }
            OutboxMessagePreview(expense)
            if (expense.lastError != null) {
                Text(
                    expense.lastError,
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.error,
                )
            }
        }
    }

    if (showDetail) {
        OutboxDetailDialog(expense = expense, onDismiss = { showDetail = false })
    }
}

/** Shows every captured field that's present, rather than picking just
 * one — title (if any) is the notification's headline, text/bigText are
 * often identical so bigText only renders when it actually differs. */
@Composable
private fun OutboxMessagePreview(expense: OutboxExpense) {
    if (expense.rawTitle == null && expense.rawText == null && expense.rawBigText == null) {
        Text("(no preview)", style = MaterialTheme.typography.bodyMedium)
        return
    }
    expense.rawTitle?.let {
        Text(it, style = MaterialTheme.typography.bodyMedium, fontWeight = FontWeight.Bold)
    }
    expense.rawText?.let {
        Text(it, style = MaterialTheme.typography.bodyMedium)
    }
    expense.rawBigText?.takeIf { it != expense.rawText }?.let {
        Text(it, style = MaterialTheme.typography.bodyMedium)
    }
}

@Composable
private fun OutboxDetailDialog(expense: OutboxExpense, onDismiss: () -> Unit) {
    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text(expense.source.uppercase()) },
        text = {
            Column {
                Text("Status: ${expense.status}", style = MaterialTheme.typography.labelMedium)
                Spacer(Modifier.height(8.dp))
                OutboxMessagePreview(expense)
                Spacer(Modifier.height(12.dp))
                DetailRow("Received", formatIsoInstant(expense.occurredAt))
                DetailRow("Captured", formatEpochMillis(expense.createdAt))
                DetailRow("Last updated", formatEpochMillis(expense.updatedAt))
                DetailRow("Attempts", expense.attemptCount.toString())
                expense.lastError?.let { DetailRow("Last error", it) }
            }
        },
        confirmButton = {
            TextButton(onClick = onDismiss) { Text("Close") }
        },
    )
}

@Composable
private fun DetailRow(label: String, value: String) {
    Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
        Text(label, style = MaterialTheme.typography.labelMedium, color = MaterialTheme.colorScheme.onSurfaceVariant)
        Text(value, style = MaterialTheme.typography.bodyMedium)
    }
}

private val displayFormatter: DateTimeFormatter =
    DateTimeFormatter.ofLocalizedDateTime(FormatStyle.MEDIUM, FormatStyle.SHORT).withZone(ZoneId.systemDefault())

/** [OutboxExpense.occurredAt] is an ISO-8601 UTC string, e.g.
 * "2026-07-26T05:20:22Z" — see TransactionNotificationListener.isoFormat. */
private fun formatIsoInstant(iso: String): String =
    try {
        displayFormatter.format(Instant.parse(iso))
    } catch (_: Exception) {
        iso
    }

private fun formatEpochMillis(millis: Long): String =
    displayFormatter.format(Instant.ofEpochMilli(millis))
