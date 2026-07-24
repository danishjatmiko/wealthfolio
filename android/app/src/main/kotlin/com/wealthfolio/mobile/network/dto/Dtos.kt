package com.wealthfolio.mobile.network.dto

import com.google.gson.annotations.SerializedName

data class LoginRequest(val email: String, val password: String)

data class UserDto(
    val id: String,
    val email: String,
    @SerializedName("display_name") val displayName: String,
    @SerializedName("avatar_url") val avatarUrl: String,
)

data class LoginResponse(
    val token: String,
    @SerializedName("expires_at") val expiresAt: String,
)

data class BudgetEnvelopeDetailDto(
    val id: String,
    val name: String,
)

data class ExpensePeriodDetailDto(
    val id: String,
    val label: String,
    val envelopes: List<BudgetEnvelopeDetailDto>,
)

data class ExpenseSourceMappingDto(
    val id: String,
    val source: String,
    @SerializedName("envelope_name") val envelopeName: String,
    @SerializedName("updated_at") val updatedAt: String,
)

data class UpsertSourceMappingRequest(@SerializedName("envelope_name") val envelopeName: String)

data class IngestExpenseRequest(
    @SerializedName("idempotency_key") val idempotencyKey: String,
    val source: String,
    @SerializedName("raw_title") val rawTitle: String?,
    @SerializedName("raw_text") val rawText: String?,
    @SerializedName("raw_big_text") val rawBigText: String?,
    @SerializedName("occurred_at") val occurredAt: String,
)

data class IngestExpenseResponse(
    val status: String,
    @SerializedName("fixed_expense_id") val fixedExpenseId: String?,
    @SerializedName("envelope_id") val envelopeId: String?,
    @SerializedName("amount_idr") val amountIdr: Long?,
    @SerializedName("merchant_name") val merchantName: String?,
)
