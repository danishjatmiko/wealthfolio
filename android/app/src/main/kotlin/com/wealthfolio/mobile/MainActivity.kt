package com.wealthfolio.mobile

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.safeDrawingPadding
import androidx.compose.ui.Modifier
import com.wealthfolio.mobile.auth.TokenStore
import com.wealthfolio.mobile.sync.SyncScheduler
import com.wealthfolio.mobile.ui.WealthfolioRoot
import dagger.hilt.android.AndroidEntryPoint
import javax.inject.Inject

@AndroidEntryPoint
class MainActivity : ComponentActivity() {

    @Inject lateinit var tokenStore: TokenStore
    @Inject lateinit var syncScheduler: SyncScheduler

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        // targetSdk 36 already forces edge-to-edge drawing on its own;
        // calling this explicitly is what additionally makes the status
        // bar/nav bar backgrounds transparent and picks correct icon
        // contrast for them, so the area around the camera cutout reads
        // as one continuous surface with the header instead of a stark
        // bar sitting above it.
        enableEdgeToEdge()

        // Belt-and-suspenders sweep (see plan Part B3) — registering this
        // is idempotent (ExistingPeriodicWorkPolicy.KEEP), so it's safe to
        // call on every app open rather than only once ever.
        syncScheduler.schedulePeriodicSweep()

        setContent {
            MaterialTheme {
                // Android 15+ (targetSdk 36 here) forces edge-to-edge
                // drawing — content renders behind the status bar/camera
                // cutout/nav bar by default. Without safeDrawingPadding,
                // the header sits underneath the cutout, visually
                // overlapping it and making its buttons effectively
                // untappable in that band. This pushes all content into
                // the actual safe area instead.
                Surface(modifier = Modifier.fillMaxSize().safeDrawingPadding()) {
                    WealthfolioRoot(tokenStore = tokenStore)
                }
            }
        }
    }
}
