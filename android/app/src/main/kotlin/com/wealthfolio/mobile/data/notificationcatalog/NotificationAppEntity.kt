package com.wealthfolio.mobile.data.notificationcatalog

import androidx.room.Entity
import androidx.room.ForeignKey
import androidx.room.Index

/**
 * Local mirror of one backend notification_apps row — the catalog of
 * e-wallet/bank apps this install is allowed to read notifications from.
 * Replaces the old hardcoded NotificationSource enum: this table is fully
 * replaced on every successful [NotificationCatalogRepository.sync] call,
 * so a new supported app appears here without an APK release.
 */
@Entity(tableName = "notification_apps")
data class NotificationAppEntity(
    @androidx.room.PrimaryKey val source: String,
    val displayName: String,
    val enabled: Boolean,
)

/**
 * Local mirror of one backend notification_app_packages row. Cascades away
 * automatically when its parent [NotificationAppEntity] is deleted, so a
 * full-replace sync never leaves orphaned package rows for a
 * removed/renamed source.
 */
@Entity(
    tableName = "notification_app_packages",
    primaryKeys = ["packageName"],
    foreignKeys = [
        ForeignKey(
            entity = NotificationAppEntity::class,
            parentColumns = ["source"],
            childColumns = ["source"],
            onDelete = ForeignKey.CASCADE,
        ),
    ],
    indices = [Index("source")],
)
data class NotificationAppPackageEntity(
    val packageName: String,
    val source: String,
)
