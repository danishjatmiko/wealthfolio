package com.wealthfolio.mobile.data.notificationcatalog

import androidx.room.Dao
import androidx.room.Insert
import androidx.room.OnConflictStrategy
import androidx.room.Query
import androidx.room.Transaction

@Dao
interface NotificationCatalogDao {
    @Query("SELECT * FROM notification_apps")
    suspend fun listApps(): List<NotificationAppEntity>

    @Query("SELECT * FROM notification_app_packages")
    suspend fun listPackages(): List<NotificationAppPackageEntity>

    @Query("DELETE FROM notification_apps")
    suspend fun deleteAllApps()

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertApps(apps: List<NotificationAppEntity>)

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertPackages(packages: List<NotificationAppPackageEntity>)

    /**
     * Full replace of the local catalog with a freshly-fetched one from
     * the backend. Deleting every app first cascades away its packages
     * (see the FK on NotificationAppPackageEntity), so a source that was
     * removed or renamed server-side never leaves orphaned package rows
     * behind.
     */
    @Transaction
    suspend fun replaceAll(apps: List<NotificationAppEntity>, packages: List<NotificationAppPackageEntity>) {
        deleteAllApps()
        insertApps(apps)
        insertPackages(packages)
    }
}
