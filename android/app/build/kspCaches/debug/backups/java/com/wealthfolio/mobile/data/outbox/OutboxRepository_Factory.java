package com.wealthfolio.mobile.data.outbox;

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
public final class OutboxRepository_Factory implements Factory<OutboxRepository> {
  private final Provider<OutboxDao> daoProvider;

  private OutboxRepository_Factory(Provider<OutboxDao> daoProvider) {
    this.daoProvider = daoProvider;
  }

  @Override
  public OutboxRepository get() {
    return newInstance(daoProvider.get());
  }

  public static OutboxRepository_Factory create(Provider<OutboxDao> daoProvider) {
    return new OutboxRepository_Factory(daoProvider);
  }

  public static OutboxRepository newInstance(OutboxDao dao) {
    return new OutboxRepository(dao);
  }
}
