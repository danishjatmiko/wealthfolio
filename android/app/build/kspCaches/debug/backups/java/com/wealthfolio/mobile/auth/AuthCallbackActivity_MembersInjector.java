package com.wealthfolio.mobile.auth;

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
public final class AuthCallbackActivity_MembersInjector implements MembersInjector<AuthCallbackActivity> {
  private final Provider<AuthRepository> authRepositoryProvider;

  private AuthCallbackActivity_MembersInjector(Provider<AuthRepository> authRepositoryProvider) {
    this.authRepositoryProvider = authRepositoryProvider;
  }

  @Override
  public void injectMembers(AuthCallbackActivity instance) {
    injectAuthRepository(instance, authRepositoryProvider.get());
  }

  public static MembersInjector<AuthCallbackActivity> create(
      Provider<AuthRepository> authRepositoryProvider) {
    return new AuthCallbackActivity_MembersInjector(authRepositoryProvider);
  }

  @InjectedFieldSignature("com.wealthfolio.mobile.auth.AuthCallbackActivity.authRepository")
  public static void injectAuthRepository(AuthCallbackActivity instance,
      AuthRepository authRepository) {
    instance.authRepository = authRepository;
  }
}
