import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // Emit a minimal standalone server bundle so the Docker image stays small.
  output: "standalone",
};

export default nextConfig;
