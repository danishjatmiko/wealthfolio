package com.wealthfolio.mobile.di;

import com.google.gson.Gson;
import dagger.internal.DaggerGenerated;
import dagger.internal.Factory;
import dagger.internal.Preconditions;
import dagger.internal.Provider;
import dagger.internal.QualifierMetadata;
import dagger.internal.ScopeMetadata;
import javax.annotation.processing.Generated;
import okhttp3.OkHttpClient;
import retrofit2.Retrofit;

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
public final class NetworkModule_ProvideRetrofitFactory implements Factory<Retrofit> {
  private final Provider<OkHttpClient> clientProvider;

  private final Provider<Gson> gsonProvider;

  private NetworkModule_ProvideRetrofitFactory(Provider<OkHttpClient> clientProvider,
      Provider<Gson> gsonProvider) {
    this.clientProvider = clientProvider;
    this.gsonProvider = gsonProvider;
  }

  @Override
  public Retrofit get() {
    return provideRetrofit(clientProvider.get(), gsonProvider.get());
  }

  public static NetworkModule_ProvideRetrofitFactory create(Provider<OkHttpClient> clientProvider,
      Provider<Gson> gsonProvider) {
    return new NetworkModule_ProvideRetrofitFactory(clientProvider, gsonProvider);
  }

  public static Retrofit provideRetrofit(OkHttpClient client, Gson gson) {
    return Preconditions.checkNotNullFromProvides(NetworkModule.INSTANCE.provideRetrofit(client, gson));
  }
}
