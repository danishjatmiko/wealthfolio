package com.wealthfolio.mobile.di;

import com.wealthfolio.mobile.data.outbox.OutboxDao;
import com.wealthfolio.mobile.data.outbox.OutboxDatabase;
import dagger.internal.DaggerGenerated;
import dagger.internal.Factory;
import dagger.internal.Preconditions;
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
public final class DatabaseModule_ProvideOutboxDaoFactory implements Factory<OutboxDao> {
  private final Provider<OutboxDatabase> databaseProvider;

  private DatabaseModule_ProvideOutboxDaoFactory(Provider<OutboxDatabase> databaseProvider) {
    this.databaseProvider = databaseProvider;
  }

  @Override
  public OutboxDao get() {
    return provideOutboxDao(databaseProvider.get());
  }

  public static DatabaseModule_ProvideOutboxDaoFactory create(
      Provider<OutboxDatabase> databaseProvider) {
    return new DatabaseModule_ProvideOutboxDaoFactory(databaseProvider);
  }

  public static OutboxDao provideOutboxDao(OutboxDatabase database) {
    return Preconditions.checkNotNullFromProvides(DatabaseModule.INSTANCE.provideOutboxDao(database));
  }
}
