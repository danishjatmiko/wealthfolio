package com.wealthfolio.mobile.auth;

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
public final class AuthInterceptor_Factory implements Factory<AuthInterceptor> {
  private final Provider<TokenStore> tokenStoreProvider;

  private AuthInterceptor_Factory(Provider<TokenStore> tokenStoreProvider) {
    this.tokenStoreProvider = tokenStoreProvider;
  }

  @Override
  public AuthInterceptor get() {
    return newInstance(tokenStoreProvider.get());
  }

  public static AuthInterceptor_Factory create(Provider<TokenStore> tokenStoreProvider) {
    return new AuthInterceptor_Factory(tokenStoreProvider);
  }

  public static AuthInterceptor newInstance(TokenStore tokenStore) {
    return new AuthInterceptor(tokenStore);
  }
}
