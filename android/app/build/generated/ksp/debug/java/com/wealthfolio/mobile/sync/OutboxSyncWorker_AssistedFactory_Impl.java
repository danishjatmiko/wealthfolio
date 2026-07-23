package com.wealthfolio.mobile.sync;

import android.content.Context;
import androidx.work.WorkerParameters;
import dagger.internal.DaggerGenerated;
import dagger.internal.InstanceFactory;
import javax.annotation.processing.Generated;
import javax.inject.Provider;

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
public final class OutboxSyncWorker_AssistedFactory_Impl implements OutboxSyncWorker_AssistedFactory {
  private final OutboxSyncWorker_Factory delegateFactory;

  OutboxSyncWorker_AssistedFactory_Impl(OutboxSyncWorker_Factory delegateFactory) {
    this.delegateFactory = delegateFactory;
  }

  @Override
  public OutboxSyncWorker create(Context p0, WorkerParameters p1) {
    return delegateFactory.get(p0, p1);
  }

  public static Provider<OutboxSyncWorker_AssistedFactory> create(
      OutboxSyncWorker_Factory delegateFactory) {
    return InstanceFactory.create(new OutboxSyncWorker_AssistedFactory_Impl(delegateFactory));
  }

  public static dagger.internal.Provider<OutboxSyncWorker_AssistedFactory> createFactoryProvider(
      OutboxSyncWorker_Factory delegateFactory) {
    return InstanceFactory.create(new OutboxSyncWorker_AssistedFactory_Impl(delegateFactory));
  }
}
