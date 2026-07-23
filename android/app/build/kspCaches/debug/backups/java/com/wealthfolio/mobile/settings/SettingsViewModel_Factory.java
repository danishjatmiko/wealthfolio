package com.wealthfolio.mobile.settings;

import com.wealthfolio.mobile.auth.AuthRepository;
import com.wealthfolio.mobile.network.ApiService;
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
public final class SettingsViewModel_Factory implements Factory<SettingsViewModel> {
  private final Provider<SourcePreferences> sourcePreferencesProvider;

  private final Provider<ApiService> apiProvider;

  private final Provider<AuthRepository> authRepositoryProvider;

  private SettingsViewModel_Factory(Provider<SourcePreferences> sourcePreferencesProvider,
      Provider<ApiService> apiProvider, Provider<AuthRepository> authRepositoryProvider) {
    this.sourcePreferencesProvider = sourcePreferencesProvider;
    this.apiProvider = apiProvider;
    this.authRepositoryProvider = authRepositoryProvider;
  }

  @Override
  public SettingsViewModel get() {
    return newInstance(sourcePreferencesProvider.get(), apiProvider.get(), authRepositoryProvider.get());
  }

  public static SettingsViewModel_Factory create(
      Provider<SourcePreferences> sourcePreferencesProvider, Provider<ApiService> apiProvider,
      Provider<AuthRepository> authRepositoryProvider) {
    return new SettingsViewModel_Factory(sourcePreferencesProvider, apiProvider, authRepositoryProvider);
  }

  public static SettingsViewModel newInstance(SourcePreferences sourcePreferences, ApiService api,
      AuthRepository authRepository) {
    return new SettingsViewModel(sourcePreferences, api, authRepository);
  }
}
