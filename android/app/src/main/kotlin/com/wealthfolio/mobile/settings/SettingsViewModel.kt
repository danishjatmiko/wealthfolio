package com.wealthfolio.mobile.settings

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.wealthfolio.mobile.auth.AuthRepository
import com.wealthfolio.mobile.data.notificationcatalog.NotificationAppEntity
import com.wealthfolio.mobile.data.notificationcatalog.NotificationCatalogRepository
import com.wealthfolio.mobile.network.ApiService
import com.wealthfolio.mobile.network.dto.UpsertSourceMappingRequest
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject
import kotlinx.coroutines.Job
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.launch

data class SourceRow(
    val source: NotificationAppEntity,
    val enabled: Boolean,
    val mappedEnvelopeName: String?,
)

data class SettingsUiState(
    val rows: List<SourceRow> = emptyList(),
    val availableEnvelopeNames: List<String> = emptyList(),
    val displayName: String? = null,
    val email: String? = null,
    val isLoading: Boolean = true,
    val error: String? = null,
)

@HiltViewModel
class SettingsViewModel @Inject constructor(
    private val sourcePreferences: SourcePreferences,
    private val catalogRepository: NotificationCatalogRepository,
    private val api: ApiService,
    private val authRepository: AuthRepository,
) : ViewModel() {

    private val _uiState = MutableStateFlow(SettingsUiState())
    val uiState: StateFlow<SettingsUiState> = _uiState.asStateFlow()

    // Cancelled and relaunched on every isLoggedIn transition below —
    // without this, this ViewModel (which MainShell's hiltViewModel()
    // calls resolve up to MainActivity's own ViewModelStore, since there's
    // no NavHost destination scoping it) would otherwise keep running the
    // *first* login's collectors forever, silently showing that account's
    // profile/mappings even after logging out and into a different one.
    private var refreshJob: Job? = null

    init {
        viewModelScope.launch {
            authRepository.isLoggedIn.collect { loggedIn ->
                refreshJob?.cancel()
                if (!loggedIn) {
                    _uiState.value = SettingsUiState()
                    return@collect
                }
                // A fresh Job per login (including the very first one,
                // since collect() immediately replays isLoggedIn's current
                // value) — refreshProfile and refreshCatalogAndMappings
                // both read whatever token is current at call time via
                // AuthInterceptor, so this always fetches the
                // just-logged-in account's own data, never a stale one.
                refreshJob = viewModelScope.launch {
                    launch { refreshProfile() }
                    refreshCatalogAndMappings()
                }
            }
        }
    }

    /** Best-effort — a failed fetch just leaves the account card blank
     * rather than blocking the notification-source UI, which is this
     * screen's actual reason for existing. */
    private suspend fun refreshProfile() {
        try {
            val user = api.me().body()
            if (user != null) {
                _uiState.value = _uiState.value.copy(displayName = user.displayName, email = user.email)
            }
        } catch (_: Exception) {
        }
    }

    /** Runs forever (until [refreshJob] is cancelled by the next login/
     * logout) reacting to per-source toggle changes, same as before —
     * only the "restart fresh on every login" wrapper around this is new. */
    private suspend fun refreshCatalogAndMappings() {
        // Best-effort — a source added server-side since the last
        // MainActivity app-open sync should still show up when the user
        // opens Settings; on failure we fall back to whatever's already
        // cached in Room from a previous sync.
        try {
            catalogRepository.sync()
        } catch (_: Exception) {
        }
        val apps = catalogRepository.listApps()

        val enabledFlows = apps.map { sourcePreferences.isEnabled(it.source) }
        combine(enabledFlows) { enabledValues -> enabledValues.toList() }
            .collect { enabledValues ->
                refreshMappingsAndEnvelopes(apps, enabledValues)
            }
    }

    private suspend fun refreshMappingsAndEnvelopes(apps: List<NotificationAppEntity>, enabledValues: List<Boolean>) {
        _uiState.value = _uiState.value.copy(isLoading = true, error = null)
        try {
            val mappingsResponse = api.listSourceMappings()
            val mappings = mappingsResponse.body().orEmpty().associateBy { it.source }

            val periodResponse = api.latestPeriod()
            val envelopeNames = periodResponse.body()?.envelopes?.map { it.name }.orEmpty()

            val rows = apps.mapIndexed { index, app ->
                SourceRow(
                    source = app,
                    enabled = enabledValues.getOrElse(index) { false },
                    mappedEnvelopeName = mappings[app.source]?.envelopeName,
                )
            }
            // .copy, not a fresh SettingsUiState — this races the profile
            // fetch above and a wholesale replace would blank out
            // displayName/email if this collector wins the race.
            _uiState.value = _uiState.value.copy(
                rows = rows,
                availableEnvelopeNames = envelopeNames,
                isLoading = false,
                error = null,
            )
        } catch (e: Exception) {
            _uiState.value = _uiState.value.copy(isLoading = false, error = e.message ?: "Failed to load settings")
        }
    }

    fun setSourceEnabled(source: NotificationAppEntity, enabled: Boolean) {
        viewModelScope.launch {
            sourcePreferences.setEnabled(source.source, enabled)
        }
    }

    fun setEnvelopeMapping(source: NotificationAppEntity, envelopeName: String) {
        viewModelScope.launch {
            try {
                api.upsertSourceMapping(source.source, UpsertSourceMappingRequest(envelopeName))
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
