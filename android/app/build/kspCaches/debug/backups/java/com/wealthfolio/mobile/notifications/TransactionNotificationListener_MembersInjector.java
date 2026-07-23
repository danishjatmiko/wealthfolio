package com.wealthfolio.mobile.notifications;

import com.wealthfolio.mobile.data.outbox.OutboxRepository;
import com.wealthfolio.mobile.settings.SourcePreferences;
import com.wealthfolio.mobile.sync.SyncScheduler;
import dagger.MembersInjector;
import dagger.internal.DaggerGenerated;
import dagger.internal.InjectedFieldSignature;
import dagger.internal.Provider;
import dagger.internal.QualifierMetadata;
import javax.annotation.processing.Generated;

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
public final class TransactionNotificationListener_MembersInjector implements MembersInjector<TransactionNotificationListener> {
  private final Provider<SourcePreferences> sourcePreferencesProvider;

  private final Provider<OutboxRepository> outboxRepositoryProvider;

  private final Provider<SyncScheduler> syncSchedulerProvider;

  private TransactionNotificationListener_MembersInjector(
      Provider<SourcePreferences> sourcePreferencesProvider,
      Provider<OutboxRepository> outboxRepositoryProvider,
      Provider<SyncScheduler> syncSchedulerProvider) {
    this.sourcePreferencesProvider = sourcePreferencesProvider;
    this.outboxRepositoryProvider = outboxRepositoryProvider;
    this.syncSchedulerProvider = syncSchedulerProvider;
  }

  @Override
  public void injectMembers(TransactionNotificationListener instance) {
    injectSourcePreferences(instance, sourcePreferencesProvider.get());
    injectOutboxRepository(instance, outboxRepositoryProvider.get());
    injectSyncScheduler(instance, syncSchedulerProvider.get());
  }

  public static MembersInjector<TransactionNotificationListener> create(
      Provider<SourcePreferences> sourcePreferencesProvider,
      Provider<OutboxRepository> outboxRepositoryProvider,
      Provider<SyncScheduler> syncSchedulerProvider) {
    return new TransactionNotificationListener_MembersInjector(sourcePreferencesProvider, outboxRepositoryProvider, syncSchedulerProvider);
  }

  @InjectedFieldSignature("com.wealthfolio.mobile.notifications.TransactionNotificationListener.sourcePreferences")
  public static void injectSourcePreferences(TransactionNotificationListener instance,
      SourcePreferences sourcePreferences) {
    instance.sourcePreferences = sourcePreferences;
  }

  @InjectedFieldSignature("com.wealthfolio.mobile.notifications.TransactionNotificationListener.outboxRepository")
  public static void injectOutboxRepository(TransactionNotificationListener instance,
      OutboxRepository outboxRepository) {
    instance.outboxRepository = outboxRepository;
  }

  @InjectedFieldSignature("com.wealthfolio.mobile.notifications.TransactionNotificationListener.syncScheduler")
  public static void injectSyncScheduler(TransactionNotificationListener instance,
      SyncScheduler syncScheduler) {
    instance.syncScheduler = syncScheduler;
  }
}
