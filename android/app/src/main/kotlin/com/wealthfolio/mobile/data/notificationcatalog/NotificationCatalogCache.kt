package com.wealthfolio.mobile.data.notificationcatalog

import javax.inject.Inject
import javax.inject.Singleton
import kotlinx.coroutines.runBlocking

/**
 * In-memory package-name -> source lookup, kept in sync with the Room
 * catalog. Exists specifically so TransactionNotificationListener can keep
 * its current behavior: a plain (non-suspend) call made before it ever
 * launches a coroutine, so notifications from every irrelevant app on the
 * phone are discarded without even scheduling work for them — a Room
 * query (suspend) can't sit on that path.
 */
@Singleton
class NotificationCatalogCache @Inject constructor(private val dao: NotificationCatalogDao) {
    @Volatile private var packageToSource: Map<String, String>? = null

    /**
     * Looks up which source (if any) owns packageName, among currently
     * enabled apps. The very first call in the process may block briefly
     * to hydrate from Room — a handful of rows, local disk only, no
     * network — every call after that is a pure in-memory map lookup.
     */
    fun sourceForPackage(packageName: String): String? =
        (packageToSource ?: hydrateBlocking())[packageName]

    private fun hydrateBlocking(): Map<String, String> = runBlocking {
        val enabledSources = dao.listApps().filter { it.enabled }.map { it.source }.toSet()
        dao.listPackages().filter { it.source in enabledSources }
            .associate { it.packageName to it.source }
            .also { packageToSource = it }
    }

    /** Called right after a [NotificationCatalogRepository.sync] writes a
     * fresh catalog to Room, so the hot path reflects it immediately
     * rather than waiting for the next lazy [hydrateBlocking]. */
    fun refresh(apps: List<NotificationAppEntity>, packages: List<NotificationAppPackageEntity>) {
        val enabledSources = apps.filter { it.enabled }.map { it.source }.toSet()
        packageToSource = packages.filter { it.source in enabledSources }
            .associate { it.packageName to it.source }
    }
}
