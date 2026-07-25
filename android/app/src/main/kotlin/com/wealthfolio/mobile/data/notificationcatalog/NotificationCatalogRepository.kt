package com.wealthfolio.mobile.data.notificationcatalog

import com.wealthfolio.mobile.network.ApiService
import javax.inject.Inject
import javax.inject.Singleton

/**
 * Syncs the notification app catalog from the backend into the local Room
 * mirror + [NotificationCatalogCache]. Called on app open (MainActivity)
 * and whenever the Settings screen opens, per the app-open/Settings-open
 * requirement — TransactionNotificationListener never calls this directly,
 * it only ever reads the already-synced cache.
 */
@Singleton
class NotificationCatalogRepository @Inject constructor(
    private val dao: NotificationCatalogDao,
    private val api: ApiService,
    private val cache: NotificationCatalogCache,
) {
    /**
     * Fetches the full catalog and replaces the local copy. Throws on any
     * failure (network, non-2xx, etc.) rather than silently wiping the
     * local catalog with an empty one — callers are expected to catch this
     * and keep whatever was cached from the last successful sync.
     */
    suspend fun sync() {
        val response = api.listNotificationApps()
        if (!response.isSuccessful) {
            error("Failed to fetch notification apps: HTTP ${response.code()}")
        }
        val dtos = response.body().orEmpty()

        val apps = dtos.map { NotificationAppEntity(source = it.source, displayName = it.displayName, enabled = it.enabled) }
        val packages = dtos.flatMap { dto ->
            dto.packageNames.map { NotificationAppPackageEntity(packageName = it, source = dto.source) }
        }

        dao.replaceAll(apps, packages)
        cache.refresh(apps, packages)
    }

    suspend fun listApps(): List<NotificationAppEntity> = dao.listApps()
}
