package com.wealthfolio.mobile.`data`.outbox

import androidx.room.EntityDeleteOrUpdateAdapter
import androidx.room.EntityInsertAdapter
import androidx.room.RoomDatabase
import androidx.room.coroutines.createFlow
import androidx.room.util.getColumnIndexOrThrow
import androidx.room.util.performSuspending
import androidx.sqlite.SQLiteStatement
import javax.`annotation`.processing.Generated
import kotlin.Int
import kotlin.Long
import kotlin.String
import kotlin.Suppress
import kotlin.Unit
import kotlin.collections.List
import kotlin.collections.MutableList
import kotlin.collections.mutableListOf
import kotlin.reflect.KClass
import kotlinx.coroutines.flow.Flow

@Generated(value = ["androidx.room.RoomProcessor"])
@Suppress(names = ["UNCHECKED_CAST", "DEPRECATION", "REDUNDANT_PROJECTION", "REMOVAL"])
public class OutboxDao_Impl(
  __db: RoomDatabase,
) : OutboxDao {
  private val __db: RoomDatabase

  private val __insertAdapterOfOutboxExpense: EntityInsertAdapter<OutboxExpense>

  private val __converters: Converters = Converters()

  private val __updateAdapterOfOutboxExpense: EntityDeleteOrUpdateAdapter<OutboxExpense>
  init {
    this.__db = __db
    this.__insertAdapterOfOutboxExpense = object : EntityInsertAdapter<OutboxExpense>() {
      protected override fun createQuery(): String = "INSERT OR IGNORE INTO `outbox_expenses` (`id`,`idempotencyKey`,`source`,`rawTitle`,`rawText`,`rawBigText`,`occurredAt`,`status`,`attemptCount`,`lastError`,`createdAt`) VALUES (nullif(?, 0),?,?,?,?,?,?,?,?,?,?)"

      protected override fun bind(statement: SQLiteStatement, entity: OutboxExpense) {
        statement.bindLong(1, entity.id)
        statement.bindText(2, entity.idempotencyKey)
        statement.bindText(3, entity.source)
        val _tmpRawTitle: String? = entity.rawTitle
        if (_tmpRawTitle == null) {
          statement.bindNull(4)
        } else {
          statement.bindText(4, _tmpRawTitle)
        }
        val _tmpRawText: String? = entity.rawText
        if (_tmpRawText == null) {
          statement.bindNull(5)
        } else {
          statement.bindText(5, _tmpRawText)
        }
        val _tmpRawBigText: String? = entity.rawBigText
        if (_tmpRawBigText == null) {
          statement.bindNull(6)
        } else {
          statement.bindText(6, _tmpRawBigText)
        }
        statement.bindText(7, entity.occurredAt)
        val _tmp: String = __converters.fromStatus(entity.status)
        statement.bindText(8, _tmp)
        statement.bindLong(9, entity.attemptCount.toLong())
        val _tmpLastError: String? = entity.lastError
        if (_tmpLastError == null) {
          statement.bindNull(10)
        } else {
          statement.bindText(10, _tmpLastError)
        }
        statement.bindLong(11, entity.createdAt)
      }
    }
    this.__updateAdapterOfOutboxExpense = object : EntityDeleteOrUpdateAdapter<OutboxExpense>() {
      protected override fun createQuery(): String = "UPDATE OR ABORT `outbox_expenses` SET `id` = ?,`idempotencyKey` = ?,`source` = ?,`rawTitle` = ?,`rawText` = ?,`rawBigText` = ?,`occurredAt` = ?,`status` = ?,`attemptCount` = ?,`lastError` = ?,`createdAt` = ? WHERE `id` = ?"

      protected override fun bind(statement: SQLiteStatement, entity: OutboxExpense) {
        statement.bindLong(1, entity.id)
        statement.bindText(2, entity.idempotencyKey)
        statement.bindText(3, entity.source)
        val _tmpRawTitle: String? = entity.rawTitle
        if (_tmpRawTitle == null) {
          statement.bindNull(4)
        } else {
          statement.bindText(4, _tmpRawTitle)
        }
        val _tmpRawText: String? = entity.rawText
        if (_tmpRawText == null) {
          statement.bindNull(5)
        } else {
          statement.bindText(5, _tmpRawText)
        }
        val _tmpRawBigText: String? = entity.rawBigText
        if (_tmpRawBigText == null) {
          statement.bindNull(6)
        } else {
          statement.bindText(6, _tmpRawBigText)
        }
        statement.bindText(7, entity.occurredAt)
        val _tmp: String = __converters.fromStatus(entity.status)
        statement.bindText(8, _tmp)
        statement.bindLong(9, entity.attemptCount.toLong())
        val _tmpLastError: String? = entity.lastError
        if (_tmpLastError == null) {
          statement.bindNull(10)
        } else {
          statement.bindText(10, _tmpLastError)
        }
        statement.bindLong(11, entity.createdAt)
        statement.bindLong(12, entity.id)
      }
    }
  }

  public override suspend fun insert(expense: OutboxExpense): Long = performSuspending(__db, false, true) { _connection ->
    val _result: Long = __insertAdapterOfOutboxExpense.insertAndReturnId(_connection, expense)
    _result
  }

  public override suspend fun update(expense: OutboxExpense): Unit = performSuspending(__db, false, true) { _connection ->
    __updateAdapterOfOutboxExpense.handle(_connection, expense)
  }

  public override suspend fun listRetryable(): List<OutboxExpense> {
    val _sql: String = "SELECT * FROM outbox_expenses WHERE status IN ('PENDING', 'FAILED') ORDER BY createdAt"
    return performSuspending(__db, true, false) { _connection ->
      val _stmt: SQLiteStatement = _connection.prepare(_sql)
      try {
        val _columnIndexOfId: Int = getColumnIndexOrThrow(_stmt, "id")
        val _columnIndexOfIdempotencyKey: Int = getColumnIndexOrThrow(_stmt, "idempotencyKey")
        val _columnIndexOfSource: Int = getColumnIndexOrThrow(_stmt, "source")
        val _columnIndexOfRawTitle: Int = getColumnIndexOrThrow(_stmt, "rawTitle")
        val _columnIndexOfRawText: Int = getColumnIndexOrThrow(_stmt, "rawText")
        val _columnIndexOfRawBigText: Int = getColumnIndexOrThrow(_stmt, "rawBigText")
        val _columnIndexOfOccurredAt: Int = getColumnIndexOrThrow(_stmt, "occurredAt")
        val _columnIndexOfStatus: Int = getColumnIndexOrThrow(_stmt, "status")
        val _columnIndexOfAttemptCount: Int = getColumnIndexOrThrow(_stmt, "attemptCount")
        val _columnIndexOfLastError: Int = getColumnIndexOrThrow(_stmt, "lastError")
        val _columnIndexOfCreatedAt: Int = getColumnIndexOrThrow(_stmt, "createdAt")
        val _result: MutableList<OutboxExpense> = mutableListOf()
        while (_stmt.step()) {
          val _item: OutboxExpense
          val _tmpId: Long
          _tmpId = _stmt.getLong(_columnIndexOfId)
          val _tmpIdempotencyKey: String
          _tmpIdempotencyKey = _stmt.getText(_columnIndexOfIdempotencyKey)
          val _tmpSource: String
          _tmpSource = _stmt.getText(_columnIndexOfSource)
          val _tmpRawTitle: String?
          if (_stmt.isNull(_columnIndexOfRawTitle)) {
            _tmpRawTitle = null
          } else {
            _tmpRawTitle = _stmt.getText(_columnIndexOfRawTitle)
          }
          val _tmpRawText: String?
          if (_stmt.isNull(_columnIndexOfRawText)) {
            _tmpRawText = null
          } else {
            _tmpRawText = _stmt.getText(_columnIndexOfRawText)
          }
          val _tmpRawBigText: String?
          if (_stmt.isNull(_columnIndexOfRawBigText)) {
            _tmpRawBigText = null
          } else {
            _tmpRawBigText = _stmt.getText(_columnIndexOfRawBigText)
          }
          val _tmpOccurredAt: String
          _tmpOccurredAt = _stmt.getText(_columnIndexOfOccurredAt)
          val _tmpStatus: OutboxStatus
          val _tmp: String
          _tmp = _stmt.getText(_columnIndexOfStatus)
          _tmpStatus = __converters.toStatus(_tmp)
          val _tmpAttemptCount: Int
          _tmpAttemptCount = _stmt.getLong(_columnIndexOfAttemptCount).toInt()
          val _tmpLastError: String?
          if (_stmt.isNull(_columnIndexOfLastError)) {
            _tmpLastError = null
          } else {
            _tmpLastError = _stmt.getText(_columnIndexOfLastError)
          }
          val _tmpCreatedAt: Long
          _tmpCreatedAt = _stmt.getLong(_columnIndexOfCreatedAt)
          _item = OutboxExpense(_tmpId,_tmpIdempotencyKey,_tmpSource,_tmpRawTitle,_tmpRawText,_tmpRawBigText,_tmpOccurredAt,_tmpStatus,_tmpAttemptCount,_tmpLastError,_tmpCreatedAt)
          _result.add(_item)
        }
        _result
      } finally {
        _stmt.close()
      }
    }
  }

  public override fun observeAll(): Flow<List<OutboxExpense>> {
    val _sql: String = "SELECT * FROM outbox_expenses ORDER BY createdAt DESC"
    return createFlow(__db, false, arrayOf("outbox_expenses")) { _connection ->
      val _stmt: SQLiteStatement = _connection.prepare(_sql)
      try {
        val _columnIndexOfId: Int = getColumnIndexOrThrow(_stmt, "id")
        val _columnIndexOfIdempotencyKey: Int = getColumnIndexOrThrow(_stmt, "idempotencyKey")
        val _columnIndexOfSource: Int = getColumnIndexOrThrow(_stmt, "source")
        val _columnIndexOfRawTitle: Int = getColumnIndexOrThrow(_stmt, "rawTitle")
        val _columnIndexOfRawText: Int = getColumnIndexOrThrow(_stmt, "rawText")
        val _columnIndexOfRawBigText: Int = getColumnIndexOrThrow(_stmt, "rawBigText")
        val _columnIndexOfOccurredAt: Int = getColumnIndexOrThrow(_stmt, "occurredAt")
        val _columnIndexOfStatus: Int = getColumnIndexOrThrow(_stmt, "status")
        val _columnIndexOfAttemptCount: Int = getColumnIndexOrThrow(_stmt, "attemptCount")
        val _columnIndexOfLastError: Int = getColumnIndexOrThrow(_stmt, "lastError")
        val _columnIndexOfCreatedAt: Int = getColumnIndexOrThrow(_stmt, "createdAt")
        val _result: MutableList<OutboxExpense> = mutableListOf()
        while (_stmt.step()) {
          val _item: OutboxExpense
          val _tmpId: Long
          _tmpId = _stmt.getLong(_columnIndexOfId)
          val _tmpIdempotencyKey: String
          _tmpIdempotencyKey = _stmt.getText(_columnIndexOfIdempotencyKey)
          val _tmpSource: String
          _tmpSource = _stmt.getText(_columnIndexOfSource)
          val _tmpRawTitle: String?
          if (_stmt.isNull(_columnIndexOfRawTitle)) {
            _tmpRawTitle = null
          } else {
            _tmpRawTitle = _stmt.getText(_columnIndexOfRawTitle)
          }
          val _tmpRawText: String?
          if (_stmt.isNull(_columnIndexOfRawText)) {
            _tmpRawText = null
          } else {
            _tmpRawText = _stmt.getText(_columnIndexOfRawText)
          }
          val _tmpRawBigText: String?
          if (_stmt.isNull(_columnIndexOfRawBigText)) {
            _tmpRawBigText = null
          } else {
            _tmpRawBigText = _stmt.getText(_columnIndexOfRawBigText)
          }
          val _tmpOccurredAt: String
          _tmpOccurredAt = _stmt.getText(_columnIndexOfOccurredAt)
          val _tmpStatus: OutboxStatus
          val _tmp: String
          _tmp = _stmt.getText(_columnIndexOfStatus)
          _tmpStatus = __converters.toStatus(_tmp)
          val _tmpAttemptCount: Int
          _tmpAttemptCount = _stmt.getLong(_columnIndexOfAttemptCount).toInt()
          val _tmpLastError: String?
          if (_stmt.isNull(_columnIndexOfLastError)) {
            _tmpLastError = null
          } else {
            _tmpLastError = _stmt.getText(_columnIndexOfLastError)
          }
          val _tmpCreatedAt: Long
          _tmpCreatedAt = _stmt.getLong(_columnIndexOfCreatedAt)
          _item = OutboxExpense(_tmpId,_tmpIdempotencyKey,_tmpSource,_tmpRawTitle,_tmpRawText,_tmpRawBigText,_tmpOccurredAt,_tmpStatus,_tmpAttemptCount,_tmpLastError,_tmpCreatedAt)
          _result.add(_item)
        }
        _result
      } finally {
        _stmt.close()
      }
    }
  }

  public override suspend fun getById(id: Long): OutboxExpense? {
    val _sql: String = "SELECT * FROM outbox_expenses WHERE id = ?"
    return performSuspending(__db, true, false) { _connection ->
      val _stmt: SQLiteStatement = _connection.prepare(_sql)
      try {
        var _argIndex: Int = 1
        _stmt.bindLong(_argIndex, id)
        val _columnIndexOfId: Int = getColumnIndexOrThrow(_stmt, "id")
        val _columnIndexOfIdempotencyKey: Int = getColumnIndexOrThrow(_stmt, "idempotencyKey")
        val _columnIndexOfSource: Int = getColumnIndexOrThrow(_stmt, "source")
        val _columnIndexOfRawTitle: Int = getColumnIndexOrThrow(_stmt, "rawTitle")
        val _columnIndexOfRawText: Int = getColumnIndexOrThrow(_stmt, "rawText")
        val _columnIndexOfRawBigText: Int = getColumnIndexOrThrow(_stmt, "rawBigText")
        val _columnIndexOfOccurredAt: Int = getColumnIndexOrThrow(_stmt, "occurredAt")
        val _columnIndexOfStatus: Int = getColumnIndexOrThrow(_stmt, "status")
        val _columnIndexOfAttemptCount: Int = getColumnIndexOrThrow(_stmt, "attemptCount")
        val _columnIndexOfLastError: Int = getColumnIndexOrThrow(_stmt, "lastError")
        val _columnIndexOfCreatedAt: Int = getColumnIndexOrThrow(_stmt, "createdAt")
        val _result: OutboxExpense?
        if (_stmt.step()) {
          val _tmpId: Long
          _tmpId = _stmt.getLong(_columnIndexOfId)
          val _tmpIdempotencyKey: String
          _tmpIdempotencyKey = _stmt.getText(_columnIndexOfIdempotencyKey)
          val _tmpSource: String
          _tmpSource = _stmt.getText(_columnIndexOfSource)
          val _tmpRawTitle: String?
          if (_stmt.isNull(_columnIndexOfRawTitle)) {
            _tmpRawTitle = null
          } else {
            _tmpRawTitle = _stmt.getText(_columnIndexOfRawTitle)
          }
          val _tmpRawText: String?
          if (_stmt.isNull(_columnIndexOfRawText)) {
            _tmpRawText = null
          } else {
            _tmpRawText = _stmt.getText(_columnIndexOfRawText)
          }
          val _tmpRawBigText: String?
          if (_stmt.isNull(_columnIndexOfRawBigText)) {
            _tmpRawBigText = null
          } else {
            _tmpRawBigText = _stmt.getText(_columnIndexOfRawBigText)
          }
          val _tmpOccurredAt: String
          _tmpOccurredAt = _stmt.getText(_columnIndexOfOccurredAt)
          val _tmpStatus: OutboxStatus
          val _tmp: String
          _tmp = _stmt.getText(_columnIndexOfStatus)
          _tmpStatus = __converters.toStatus(_tmp)
          val _tmpAttemptCount: Int
          _tmpAttemptCount = _stmt.getLong(_columnIndexOfAttemptCount).toInt()
          val _tmpLastError: String?
          if (_stmt.isNull(_columnIndexOfLastError)) {
            _tmpLastError = null
          } else {
            _tmpLastError = _stmt.getText(_columnIndexOfLastError)
          }
          val _tmpCreatedAt: Long
          _tmpCreatedAt = _stmt.getLong(_columnIndexOfCreatedAt)
          _result = OutboxExpense(_tmpId,_tmpIdempotencyKey,_tmpSource,_tmpRawTitle,_tmpRawText,_tmpRawBigText,_tmpOccurredAt,_tmpStatus,_tmpAttemptCount,_tmpLastError,_tmpCreatedAt)
        } else {
          _result = null
        }
        _result
      } finally {
        _stmt.close()
      }
    }
  }

  public companion object {
    public fun getRequiredConverters(): List<KClass<*>> = emptyList()
  }
}
