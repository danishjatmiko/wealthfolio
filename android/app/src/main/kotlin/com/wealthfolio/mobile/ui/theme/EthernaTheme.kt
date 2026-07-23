package com.wealthfolio.mobile.ui.theme

import androidx.compose.material3.lightColorScheme
import androidx.compose.ui.graphics.Color

/** Mirrors the web app's brand tokens (frontend/src/styles/tokens.css
 * --accent/--accent-gold) so the native screens match the logo/launcher
 * icon instead of falling back to Material3's default purple baseline. */
val EthernaForest = Color(0xFF1F3A2A)
val EthernaForestDark = Color(0xFF16291D)
val EthernaGold = Color(0xFFC08A2E)
val EthernaRed = Color(0xFFB3402F)
val EthernaGreenSoft = Color(0xFFE4F1E8)
val EthernaGreenText = Color(0xFF2F6B46)
val EthernaGoldSoft = Color(0xFFF3E7D2)
val EthernaGoldText = Color(0xFF8A611F)
val EthernaRedSoft = Color(0xFFF6E4E0)

val EthernaColorScheme = lightColorScheme(
    primary = EthernaForest,
    onPrimary = Color.White,
    primaryContainer = EthernaGreenSoft,
    onPrimaryContainer = EthernaForestDark,
    secondary = EthernaGold,
    onSecondary = Color.White,
    error = EthernaRed,
    onError = Color.White,
)
