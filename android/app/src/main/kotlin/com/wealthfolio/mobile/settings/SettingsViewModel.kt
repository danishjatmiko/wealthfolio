package com.wealthfolio.mobile.settings

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.wealthfolio.mobile.auth.AuthRepository
import com.wealthfolio.mobile.network.ApiService
import com.wealthfolio.mobile.network.dto.UpsertSourceMappingRequest
import com.wealthfolio.mobile.notifications.NotificationSource
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.launch

data class SourceRow(
    val source: NotificationSource,
    val enabled: Boolean,
    val mappedEnvelopeName: String?,
)

data class SettingsUiState(
    val rows: List<SourceRow> = emptyList(),
    val availableEnvelopeNames: List<String> = emptyList(),
    val isLoading: Boolean = true,
    val error: String? = null,
)

@HiltViewModel
class SettingsViewModel @Inject constructor(
    private val sourcePreferences: SourcePreferences,
    private val api: ApiService,
    private val authRepository: AuthRepository,
) : ViewModel() {

    private val _uiState = MutableStateFlow(SettingsUiState())
    val uiState: StateFlow<SettingsUiState> = _uiState.asStateFlow()

    init {
        viewModelScope.launch {
            val enabledFlows = NotificationSource.entries.map { source ->
                sourcePreferences.isEnabled(source)
            }
            combine(enabledFlows) { enabledValues -> enabledValues.toList() }
                .collect { enabledValues ->
                    refreshMappingsAndEnvelopes(enabledValues)
                }
        }
    }

    private suspend fun refreshMappingsAndEnvelopes(enabledValues: List<Boolean>) {
        _uiState.value = _uiState.value.copy(isLoading = true, error = null)
        try {
            val mappingsResponse = api.listSourceMappings()
            val mappings = mappingsResponse.body().orEmpty().associateBy { it.source }

            val periodResponse = api.latestPeriod()
            val envelopeNames = periodResponse.body()?.envelopes?.map { it.name }.orEmpty()

            val rows = NotificationSource.entries.mapIndexed { index, source ->
                SourceRow(
                    source = source,
                    enabled = enabledValues.getOrElse(index) { false },
                    mappedEnvelopeName = mappings[source.id]?.envelopeName,
                )
            }
            _uiState.value = SettingsUiState(rows = rows, availableEnvelopeNames = envelopeNames, isLoading = false)
        } catch (e: Exception) {
            _uiState.value = _uiState.value.copy(isLoading = false, error = e.message ?: "Failed to load settings")
        }
    }

    fun setSourceEnabled(source: NotificationSource, enabled: Boolean) {
        viewModelScope.launch {
            sourcePreferences.setEnabled(source, enabled)
        }
    }

    fun setEnvelopeMapping(source: NotificationSource, envelopeName: String) {
        viewModelScope.launch {
            try {
                api.upsertSourceMapping(source.id, UpsertSourceMappingRequest(envelopeName))
                val rows = _uiState.value.rows.map {
                    if (it.source == source) it.copy(mappedEnvelopeName = envelopeName) else it
                }
                _uiState.value = _uiState.value.copy(rows = rows)
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(error = e.message ?: "Failed to save mapping")
            }
        }
    }

    /** Clearing TokenStore (inside authRepository.logout()) flips
     * WealthfolioRoot's isLoggedIn StateFlow to false, which switches the
     * whole app back to LoginScreen on its own — no navigation call
     * needed here. */
    fun logout() {
        viewModelScope.launch {
            authRepository.logout()
        }
    }
}
