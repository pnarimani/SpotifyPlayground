import { useEffect, useState } from "react";
import { Platform, Pressable, Text, View } from "react-native";
import * as SecureStore from "expo-secure-store";

const USER_ID_KEY = "user_id";
const JWT_KEY = "jwt";

const storage = {
  get: (key: string) =>
    Platform.OS === "web"
      ? Promise.resolve(localStorage.getItem(key))
      : SecureStore.getItemAsync(key),
  set: (key: string, value: string) =>
    Platform.OS === "web"
      ? Promise.resolve(localStorage.setItem(key, value))
      : SecureStore.setItemAsync(key, value),
};
const SIGNUP_URL =
  process.env.EXPO_PUBLIC_AUTH_SIGNUP_URL ??
  "http://localhost:80/api/1/auth/signup";

type SignupResponse = {
  user_id: string;
  jwt: string;
};

export default function Index() {
  const [userId, setUserId] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [signingUp, setSigningUp] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const loadStoredAuth = async () => {
      try {
        const [storedUserId, storedJwt] = await Promise.all([
          storage.get(USER_ID_KEY),
          storage.get(JWT_KEY),
        ]);

        if (storedUserId && storedJwt) {
          setUserId(storedUserId);
        }
      } finally {
        setLoading(false);
      }
    };

    loadStoredAuth();
  }, []);

  const signup = async () => {
    if (userId || signingUp || loading) {
      return;
    }

    setSigningUp(true);
    setError(null);

    try {
      const response = await fetch(SIGNUP_URL, { method: "GET" });
      if (!response.ok) {
        throw new Error(`Signup failed: ${response.status}`);
      }

      const body = (await response.json()) as SignupResponse;

      await Promise.all([
        storage.set(USER_ID_KEY, body.user_id),
        storage.set(JWT_KEY, body.jwt),
      ]);

      setUserId(body.user_id);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Signup failed");
    } finally {
      setSigningUp(false);
    }
  };

  const isSignedUp = Boolean(userId);

  return (
    <View
      style={{
        flex: 1,
        justifyContent: "center",
        alignItems: "center",
        padding: 24,
        gap: 16,
      }}
    >
      <Pressable
        onPress={signup}
        disabled={isSignedUp || signingUp || loading}
        style={{
          paddingVertical: 12,
          paddingHorizontal: 20,
          borderRadius: 8,
          backgroundColor: isSignedUp ? "#999" : "#1DB954",
        }}
      >
        <Text style={{ color: "white", fontWeight: "600" }}>
          {isSignedUp ? "Signed up" : signingUp ? "Signing up..." : "Sign up"}
        </Text>
      </Pressable>

      {userId ? <Text>User ID: {userId}</Text> : null}
      {error ? <Text style={{ color: "red" }}>{error}</Text> : null}
    </View>
  );
}
