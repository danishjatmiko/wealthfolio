package com.wealthfolio.mobile.sync

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.wealthfolio.mobile.data.outbox.OutboxExpense
import com.wealthfolio.mobile.data.outbox.OutboxRepository
import com.wealthfolio.mobile.data.outbox.OutboxStatus
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.launch

data class SyncStatusUiState(
    val pending: List<OutboxExpense> = emptyList(),
    val failed: List<OutboxExpense> = emptyList(),
    val ignored: List<OutboxExpense> = emptyList(),
    val sent: List<OutboxExpense> = emptyList(),
    val autoClearEnabled: Boolean = false,
)

@HiltViewModel
class SyncStatusViewModel @Inject constructor(
    private val outboxRepository: OutboxRepository,
    private val syncScheduler: SyncScheduler,
    private val syncPreferences: SyncPreferences,
) : ViewModel() {

    val uiState: StateFlow<SyncStatusUiState> =
        combine(outboxRepository.observeAll(), syncPreferences.autoClearEnabled) { all, autoClearEnabled ->
            SyncStatusUiState(
                pending = all.filter { it.status == OutboxStatus.PENDING },
                failed = all.filter { it.status == OutboxStatus.FAILED },
                ignored = all.filter { it.status == OutboxStatus.IGNORED },
                sent = all.filter { it.status == OutboxStatus.SENT },
                autoClearEnabled = autoClearEnabled,
            )
        }
            .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5_000), SyncStatusUiState())

    // Outbox data itself is already live (Room Flow above), so "refresh"
    // here means "go attempt to sync pending/failed items right now"
    // rather than re-reading anything. syncNow() is a fire-and-forget
    // WorkManager enqueue with no easy completion signal, so this just
    // holds the pull-to-refresh spinner up briefly as acknowledgment
    // rather than tracking the work's actual finish.
    private val _isRefreshing = MutableStateFlow(false)
    val isRefreshing: StateFlow<Boolean> = _isRefreshing.asStateFlow()

    fun refresh() {
        viewModelScope.launch {
            _isRefreshing.value = true
            syncScheduler.syncNow()
            delay(600)
            _isRefreshing.value = false
        }
    }

    fun retry(expense: OutboxExpense) {
        viewModelScope.launch {
            outboxRepository.requeue(expense.id)
            syncScheduler.syncNow()
        }
    }

    /** Toggling on immediately clears resolved history (the "when turned
     * on, remove the data" ask) and starts the recurring daily job;
     * toggling off just cancels the job going forward. */
    fun setAutoClearEnabled(enabled: Boolean) {
        viewModelScope.launch {
            syncPreferences.setAutoClearEnabled(enabled)
            if (enabled) {
                outboxRepository.clearResolved()
                syncScheduler.scheduleAutoClear()
            } else {
                syncScheduler.cancelAutoClear()
            }
        }
    }

    fun clearResolved() {
        viewModelScope.launch { outboxRepository.clearResolved() }
    }

    fun clearAll() {
        viewModelScope.launch { outboxRepository.clearAll() }
    }
}
