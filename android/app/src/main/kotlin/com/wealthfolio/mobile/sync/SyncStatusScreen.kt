package com.wealthfolio.mobile.sync

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
import androidx.compose.material3.Card
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.wealthfolio.mobile.data.outbox.OutboxExpense
import com.wealthfolio.mobile.ui.NativeScreenTopBar

@Composable
fun SyncStatusScreen(onBack: () -> Unit, viewModel: SyncStatusViewModel = hiltViewModel()) {
    val state by viewModel.uiState.collectAsState()

    Scaffold(topBar = { NativeScreenTopBar("Sync status", onBack) }) { padding ->
        Column(modifier = Modifier.fillMaxSize().padding(padding).padding(16.dp)) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
            ) {
                Text("Outbox", style = MaterialTheme.typography.titleLarge)
                OutlinedButton(onClick = viewModel::retryAll) {
                    Text("Retry all")
                }
            }
            Spacer(Modifier.height(12.dp))

            LazyColumn {
                section("Pending", state.pending) { null }
                section("Failed", state.failed) { expense -> { viewModel.retry(expense) } }
                section("Ignored (not a recognized transaction)", state.ignored) { null }
                section("Recently synced", state.sent) { null }
            }
        }
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
    Card(modifier = Modifier.fillMaxWidth()) {
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
            Text(expense.rawTitle ?: expense.rawText ?: "(no preview)", style = MaterialTheme.typography.bodyMedium)
            if (expense.lastError != null) {
                Text(
                    expense.lastError,
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.error,
                )
            }
        }
    }
}
