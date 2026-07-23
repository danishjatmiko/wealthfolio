package com.wealthfolio.mobile.sync;

import com.wealthfolio.mobile.data.outbox.OutboxRepository;
import dagger.internal.DaggerGenerated;
import dagger.internal.Factory;
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
public final class SyncStatusViewModel_Factory implements Factory<SyncStatusViewModel> {
  private final Provider<OutboxRepository> outboxRepositoryProvider;

  private final Provider<SyncScheduler> syncSchedulerProvider;

  private SyncStatusViewModel_Factory(Provider<OutboxRepository> outboxRepositoryProvider,
      Provider<SyncScheduler> syncSchedulerProvider) {
    this.outboxRepositoryProvider = outboxRepositoryProvider;
    this.syncSchedulerProvider = syncSchedulerProvider;
  }

  @Override
  public SyncStatusViewModel get() {
    return newInstance(outboxRepositoryProvider.get(), syncSchedulerProvider.get());
  }

  public static SyncStatusViewModel_Factory create(
      Provider<OutboxRepository> outboxRepositoryProvider,
      Provider<SyncScheduler> syncSchedulerProvider) {
    return new SyncStatusViewModel_Factory(outboxRepositoryProvider, syncSchedulerProvider);
  }

  public static SyncStatusViewModel newInstance(OutboxRepository outboxRepository,
      SyncScheduler syncScheduler) {
    return new SyncStatusViewModel(outboxRepository, syncScheduler);
  }
}
