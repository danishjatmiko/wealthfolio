package com.wealthfolio.mobile.sync

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.wealthfolio.mobile.data.outbox.OutboxExpense
import com.wealthfolio.mobile.data.outbox.OutboxRepository
import com.wealthfolio.mobile.data.outbox.OutboxStatus
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.launch

data class SyncStatusUiState(
    val pending: List<OutboxExpense> = emptyList(),
    val failed: List<OutboxExpense> = emptyList(),
    val ignored: List<OutboxExpense> = emptyList(),
    val sent: List<OutboxExpense> = emptyList(),
)

@HiltViewModel
class SyncStatusViewModel @Inject constructor(
    private val outboxRepository: OutboxRepository,
    private val syncScheduler: SyncScheduler,
) : ViewModel() {

    val uiState: StateFlow<SyncStatusUiState> = outboxRepository.observeAll()
        .map { all ->
            SyncStatusUiState(
                pending = all.filter { it.status == OutboxStatus.PENDING },
                failed = all.filter { it.status == OutboxStatus.FAILED },
                ignored = all.filter { it.status == OutboxStatus.IGNORED },
                sent = all.filter { it.status == OutboxStatus.SENT },
            )
        }
        .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5_000), SyncStatusUiState())

    fun retryAll() {
        syncScheduler.syncNow()
    }

    fun retry(expense: OutboxExpense) {
        viewModelScope.launch {
            outboxRepository.requeue(expense.id)
            syncScheduler.syncNow()
        }
    }
}
