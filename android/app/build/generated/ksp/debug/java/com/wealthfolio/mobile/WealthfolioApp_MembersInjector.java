package com.wealthfolio.mobile;

import androidx.hilt.work.HiltWorkerFactory;
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
public final class WealthfolioApp_MembersInjector implements MembersInjector<WealthfolioApp> {
  private final Provider<HiltWorkerFactory> workerFactoryProvider;

  private WealthfolioApp_MembersInjector(Provider<HiltWorkerFactory> workerFactoryProvider) {
    this.workerFactoryProvider = workerFactoryProvider;
  }

  @Override
  public void injectMembers(WealthfolioApp instance) {
    injectWorkerFactory(instance, workerFactoryProvider.get());
  }

  public static MembersInjector<WealthfolioApp> create(
      Provider<HiltWorkerFactory> workerFactoryProvider) {
    return new WealthfolioApp_MembersInjector(workerFactoryProvider);
  }

  @InjectedFieldSignature("com.wealthfolio.mobile.WealthfolioApp.workerFactory")
  public static void injectWorkerFactory(WealthfolioApp instance, HiltWorkerFactory workerFactory) {
    instance.workerFactory = workerFactory;
  }
}
