package com.wealthfolio.mobile.di;

import android.content.Context;
import com.wealthfolio.mobile.data.outbox.OutboxDatabase;
import dagger.internal.DaggerGenerated;
import dagger.internal.Factory;
import dagger.internal.Preconditions;
import dagger.internal.Provider;
import dagger.internal.QualifierMetadata;
import dagger.internal.ScopeMetadata;
import javax.annotation.processing.Generated;

@ScopeMetadata("javax.inject.Singleton")
@QualifierMetadata("dagger.hilt.android.qualifiers.ApplicationContext")
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
public final class DatabaseModule_ProvideOutboxDatabaseFactory implements Factory<OutboxDatabase> {
  private final Provider<Context> contextProvider;

  private DatabaseModule_ProvideOutboxDatabaseFactory(Provider<Context> contextProvider) {
    this.contextProvider = contextProvider;
  }

  @Override
  public OutboxDatabase get() {
    return provideOutboxDatabase(contextProvider.get());
  }

  public static DatabaseModule_ProvideOutboxDatabaseFactory create(
      Provider<Context> contextProvider) {
    return new DatabaseModule_ProvideOutboxDatabaseFactory(contextProvider);
  }

  public static OutboxDatabase provideOutboxDatabase(Context context) {
    return Preconditions.checkNotNullFromProvides(DatabaseModule.INSTANCE.provideOutboxDatabase(context));
  }
}
