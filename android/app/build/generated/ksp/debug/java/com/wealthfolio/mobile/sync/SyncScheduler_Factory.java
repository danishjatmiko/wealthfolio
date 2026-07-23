package com.wealthfolio.mobile.sync;

import androidx.work.WorkManager;
import dagger.internal.DaggerGenerated;
import dagger.internal.Factory;
import dagger.internal.Provider;
import dagger.internal.QualifierMetadata;
import dagger.internal.ScopeMetadata;
import javax.annotation.processing.Generated;

@ScopeMetadata("javax.inject.Singleton")
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
public final class SyncScheduler_Factory implements Factory<SyncScheduler> {
  private final Provider<WorkManager> workManagerProvider;

  private SyncScheduler_Factory(Provider<WorkManager> workManagerProvider) {
    this.workManagerProvider = workManagerProvider;
  }

  @Override
  public SyncScheduler get() {
    return newInstance(workManagerProvider.get());
  }

  public static SyncScheduler_Factory create(Provider<WorkManager> workManagerProvider) {
    return new SyncScheduler_Factory(workManagerProvider);
  }

  public static SyncScheduler newInstance(WorkManager workManager) {
    return new SyncScheduler(workManager);
  }
}
