package com.wealthfolio.mobile.`data`.outbox

import androidx.room.InvalidationTracker
import androidx.room.RoomOpenDelegate
import androidx.room.migration.AutoMigrationSpec
import androidx.room.migration.Migration
import androidx.room.util.TableInfo
import androidx.room.util.TableInfo.Companion.read
import androidx.room.util.dropFtsSyncTriggers
import androidx.sqlite.SQLiteConnection
import androidx.sqlite.execSQL
import javax.`annotation`.processing.Generated
import kotlin.Lazy
import kotlin.String
import kotlin.Suppress
import kotlin.collections.List
import kotlin.collections.Map
import kotlin.collections.MutableList
import kotlin.collections.MutableMap
import kotlin.collections.MutableSet
import kotlin.collections.Set
import kotlin.collections.mutableListOf
import kotlin.collections.mutableMapOf
import kotlin.collections.mutableSetOf
import kotlin.reflect.KClass

@Generated(value = ["androidx.room.RoomProcessor"])
@Suppress(names = ["UNCHECKED_CAST", "DEPRECATION", "REDUNDANT_PROJECTION", "REMOVAL"])
public class OutboxDatabase_Impl : OutboxDatabase() {
  private val _outboxDao: Lazy<OutboxDao> = lazy {
    OutboxDao_Impl(this)
  }

  protected override fun createOpenDelegate(): RoomOpenDelegate {
    val _openDelegate: RoomOpenDelegate = object : RoomOpenDelegate(1, "e3b8f54977a06db4b1febcc1fb0431ca", "7e7a4780930ec9bd83a33260299ff078") {
      public override fun createAllTables(connection: SQLiteConnection) {
        connection.execSQL("CREATE TABLE IF NOT EXISTS `outbox_expenses` (`id` INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, `idempotencyKey` TEXT NOT NULL, `source` TEXT NOT NULL, `rawTitle` TEXT, `rawText` TEXT, `rawBigText` TEXT, `occurredAt` TEXT NOT NULL, `status` TEXT NOT NULL, `attemptCount` INTEGER NOT NULL, `lastError` TEXT, `createdAt` INTEGER NOT NULL)")
        connection.execSQL("CREATE UNIQUE INDEX IF NOT EXISTS `index_outbox_expenses_idempotencyKey` ON `outbox_expenses` (`idempotencyKey`)")
        connection.execSQL("CREATE TABLE IF NOT EXISTS room_master_table (id INTEGER PRIMARY KEY,identity_hash TEXT)")
        connection.execSQL("INSERT OR REPLACE INTO room_master_table (id,identity_hash) VALUES(42, 'e3b8f54977a06db4b1febcc1fb0431ca')")
      }

      public override fun dropAllTables(connection: SQLiteConnection) {
        connection.execSQL("DROP TABLE IF EXISTS `outbox_expenses`")
      }

      public override fun onCreate(connection: SQLiteConnection) {
      }

      public override fun onOpen(connection: SQLiteConnection) {
        internalInitInvalidationTracker(connection)
      }

      public override fun onPreMigrate(connection: SQLiteConnection) {
        dropFtsSyncTriggers(connection)
      }

      public override fun onPostMigrate(connection: SQLiteConnection) {
      }

      public override fun onValidateSchema(connection: SQLiteConnection): RoomOpenDelegate.ValidationResult {
        val _columnsOutboxExpenses: MutableMap<String, TableInfo.Column> = mutableMapOf()
        _columnsOutboxExpenses.put("id", TableInfo.Column("id", "INTEGER", true, 1, null, TableInfo.CREATED_FROM_ENTITY))
        _columnsOutboxExpenses.put("idempotencyKey", TableInfo.Column("idempotencyKey", "TEXT", true, 0, null, TableInfo.CREATED_FROM_ENTITY))
        _columnsOutboxExpenses.put("source", TableInfo.Column("source", "TEXT", true, 0, null, TableInfo.CREATED_FROM_ENTITY))
        _columnsOutboxExpenses.put("rawTitle", TableInfo.Column("rawTitle", "TEXT", false, 0, null, TableInfo.CREATED_FROM_ENTITY))
        _columnsOutboxExpenses.put("rawText", TableInfo.Column("rawText", "TEXT", false, 0, null, TableInfo.CREATED_FROM_ENTITY))
        _columnsOutboxExpenses.put("rawBigText", TableInfo.Column("rawBigText", "TEXT", false, 0, null, TableInfo.CREATED_FROM_ENTITY))
        _columnsOutboxExpenses.put("occurredAt", TableInfo.Column("occurredAt", "TEXT", true, 0, null, TableInfo.CREATED_FROM_ENTITY))
        _columnsOutboxExpenses.put("status", TableInfo.Column("status", "TEXT", true, 0, null, TableInfo.CREATED_FROM_ENTITY))
        _columnsOutboxExpenses.put("attemptCount", TableInfo.Column("attemptCount", "INTEGER", true, 0, null, TableInfo.CREATED_FROM_ENTITY))
        _columnsOutboxExpenses.put("lastError", TableInfo.Column("lastError", "TEXT", false, 0, null, TableInfo.CREATED_FROM_ENTITY))
        _columnsOutboxExpenses.put("createdAt", TableInfo.Column("createdAt", "INTEGER", true, 0, null, TableInfo.CREATED_FROM_ENTITY))
        val _foreignKeysOutboxExpenses: MutableSet<TableInfo.ForeignKey> = mutableSetOf()
        val _indicesOutboxExpenses: MutableSet<TableInfo.Index> = mutableSetOf()
        _indicesOutboxExpenses.add(TableInfo.Index("index_outbox_expenses_idempotencyKey", true, listOf("idempotencyKey"), listOf("ASC")))
        val _infoOutboxExpenses: TableInfo = TableInfo("outbox_expenses", _columnsOutboxExpenses, _foreignKeysOutboxExpenses, _indicesOutboxExpenses)
        val _existingOutboxExpenses: TableInfo = read(connection, "outbox_expenses")
        if (!_infoOutboxExpenses.equals(_existingOutboxExpenses)) {
          return RoomOpenDelegate.ValidationResult(false, """
              |outbox_expenses(com.wealthfolio.mobile.data.outbox.OutboxExpense).
              | Expected:
              |""".trimMargin() + _infoOutboxExpenses + """
              |
              | Found:
              |""".trimMargin() + _existingOutboxExpenses)
        }
        return RoomOpenDelegate.ValidationResult(true, null)
      }
    }
    return _openDelegate
  }

  protected override fun createInvalidationTracker(): InvalidationTracker {
    val _shadowTablesMap: MutableMap<String, String> = mutableMapOf()
    val _viewTables: MutableMap<String, Set<String>> = mutableMapOf()
    return InvalidationTracker(this, _shadowTablesMap, _viewTables, "outbox_expenses")
  }

  public override fun clearAllTables() {
    super.performClear(false, "outbox_expenses")
  }

  protected override fun getRequiredTypeConverterClasses(): Map<KClass<*>, List<KClass<*>>> {
    val _typeConvertersMap: MutableMap<KClass<*>, List<KClass<*>>> = mutableMapOf()
    _typeConvertersMap.put(OutboxDao::class, OutboxDao_Impl.getRequiredConverters())
    return _typeConvertersMap
  }

  public override fun getRequiredAutoMigrationSpecClasses(): Set<KClass<out AutoMigrationSpec>> {
    val _autoMigrationSpecsSet: MutableSet<KClass<out AutoMigrationSpec>> = mutableSetOf()
    return _autoMigrationSpecsSet
  }

  public override fun createAutoMigrations(autoMigrationSpecs: Map<KClass<out AutoMigrationSpec>, AutoMigrationSpec>): List<Migration> {
    val _autoMigrations: MutableList<Migration> = mutableListOf()
    return _autoMigrations
  }

  public override fun outboxDao(): OutboxDao = _outboxDao.value
}
