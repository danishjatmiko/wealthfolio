package com.wealthfolio.mobile;

import com.wealthfolio.mobile.auth.TokenStore;
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
public final class MainActivity_MembersInjector implements MembersInjector<MainActivity> {
  private final Provider<TokenStore> tokenStoreProvider;

  private final Provider<SyncScheduler> syncSchedulerProvider;

  private MainActivity_MembersInjector(Provider<TokenStore> tokenStoreProvider,
      Provider<SyncScheduler> syncSchedulerProvider) {
    this.tokenStoreProvider = tokenStoreProvider;
    this.syncSchedulerProvider = syncSchedulerProvider;
  }

  @Override
  public void injectMembers(MainActivity instance) {
    injectTokenStore(instance, tokenStoreProvider.get());
    injectSyncScheduler(instance, syncSchedulerProvider.get());
  }

  public static MembersInjector<MainActivity> create(Provider<TokenStore> tokenStoreProvider,
      Provider<SyncScheduler> syncSchedulerProvider) {
    return new MainActivity_MembersInjector(tokenStoreProvider, syncSchedulerProvider);
  }

  @InjectedFieldSignature("com.wealthfolio.mobile.MainActivity.tokenStore")
  public static void injectTokenStore(MainActivity instance, TokenStore tokenStore) {
    instance.tokenStore = tokenStore;
  }

  @InjectedFieldSignature("com.wealthfolio.mobile.MainActivity.syncScheduler")
  public static void injectSyncScheduler(MainActivity instance, SyncScheduler syncScheduler) {
    instance.syncScheduler = syncScheduler;
  }
}
