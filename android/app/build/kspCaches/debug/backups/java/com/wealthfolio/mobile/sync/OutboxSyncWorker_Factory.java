package com.wealthfolio.mobile.sync;

import android.content.Context;
import androidx.work.WorkerParameters;
import com.wealthfolio.mobile.data.outbox.OutboxRepository;
import com.wealthfolio.mobile.network.ApiService;
import dagger.internal.DaggerGenerated;
import dagger.internal.Provider;
import dagger.internal.QualifierMetadata;
import dagger.internal.ScopeMetadata;
import javax.annotation.processing.Generated;

@ScopeMetadata
@QualifierMetadata
@DaggerGenerated
@Generated(
    value = "dagger.internal.codegen.ComponentProcessor",
    comments = "https://dagger.dev"
)
@SuppressWarnings({
    "unchecked",
    "rawtypes",
    "KotlinInternal",
    "KotlinInternalInJava",
    "cast",
    "deprecation",
    "nullness:initialization.field.uninitialized"
})
public final class OutboxSyncWorker_Factory {
  private final Provider<OutboxRepository> outboxRepositoryProvider;

  private final Provider<ApiService> apiProvider;

  private OutboxSyncWorker_Factory(Provider<OutboxRepository> outboxRepositoryProvider,
      Provider<ApiService> apiProvider) {
    this.outboxRepositoryProvider = outboxRepositoryProvider;
    this.apiProvider = apiProvider;
  }

  public OutboxSyncWorker get(Context context, WorkerParameters params) {
    return newInstance(context, params, outboxRepositoryProvider.get(), apiProvider.get());
  }

  public static OutboxSyncWorker_Factory create(Provider<OutboxRepository> outboxRepositoryProvider,
      Provider<ApiService> apiProvider) {
    return new OutboxSyncWorker_Factory(outboxRepositoryProvider, apiProvider);
  }

  public static OutboxSyncWorker newInstance(Context context, WorkerParameters params,
      OutboxRepository outboxRepository, ApiService api) {
    return new OutboxSyncWorker(context, params, outboxRepository, api);
  }
}
