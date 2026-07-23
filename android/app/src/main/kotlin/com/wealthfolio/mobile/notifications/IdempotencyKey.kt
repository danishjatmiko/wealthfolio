package com.wealthfolio.mobile.notifications

import java.security.MessageDigest

/**
 * Deterministic key for a captured notification, computed once at capture
 * time and persisted with the outbox row — every retry of that row must
 * reuse the identical key regardless of how many attempts it takes, since
 * that's what lets the backend's UNIQUE(user_id, idempotency_key)
 * constraint dedupe correctly. Built from raw notification identity, not
 * anything parsed (parsing happens server-side — see notificationparse).
 * postTime is rounded to the minute so the key doesn't shift if the same
 * logical notification gets reposted a few milliseconds apart.
 */
fun buildIdempotencyKey(source: String, packageName: String, title: String?, text: String?, postTimeMillis: Long): String {
    val roundedMinute = postTimeMillis / 60_000
    val raw = "$source:$packageName:${title.orEmpty()}:${text.orEmpty()}:$roundedMinute"
    val digest = MessageDigest.getInstance("SHA-256").digest(raw.toByteArray(Charsets.UTF_8))
    return digest.joinToString("") { "%02x".format(it) }
}
