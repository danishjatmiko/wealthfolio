package com.wealthfolio.mobile.network

import com.wealthfolio.mobile.network.dto.ExpensePeriodDetailDto
import com.wealthfolio.mobile.network.dto.ExpenseSourceMappingDto
import com.wealthfolio.mobile.network.dto.IngestExpenseRequest
import com.wealthfolio.mobile.network.dto.IngestExpenseResponse
import com.wealthfolio.mobile.network.dto.LoginRequest
import com.wealthfolio.mobile.network.dto.LoginResponse
import com.wealthfolio.mobile.network.dto.UpsertSourceMappingRequest
import com.wealthfolio.mobile.network.dto.UserDto
import retrofit2.Response
import retrofit2.http.Body
import retrofit2.http.GET
import retrofit2.http.POST
import retrofit2.http.PUT
import retrofit2.http.Path

interface ApiService {
    @POST("auth/login")
    suspend fun login(@Body req: LoginRequest): Response<LoginResponse>

    @POST("auth/logout")
    suspend fun logout(): Response<Unit>

    @GET("auth/me")
    suspend fun me(): Response<UserDto>

    @GET("expense-periods/latest")
    suspend fun latestPeriod(): Response<ExpensePeriodDetailDto>

    @GET("expense-source-mappings")
    suspend fun listSourceMappings(): Response<List<ExpenseSourceMappingDto>>

    @PUT("expense-source-mappings/{source}")
    suspend fun upsertSourceMapping(
        @Path("source") source: String,
        @Body req: UpsertSourceMappingRequest,
    ): Response<ExpenseSourceMappingDto>

    @POST("expense-ingestions")
    suspend fun ingestExpense(@Body req: IngestExpenseRequest): Response<IngestExpenseResponse>
}
