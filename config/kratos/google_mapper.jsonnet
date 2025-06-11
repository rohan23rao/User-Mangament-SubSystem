// config/kratos/google_mapper.jsonnet
local claims = {
  email_verified: false
} + std.extVar('claims');

{
  identity: {
    traits: {
      email: claims.email,
      name: {
        first: claims.given_name,
        last: claims.family_name,
      },
    },
    // Mark Google OAuth users as verified
    verifiable_addresses: [
      {
        value: claims.email,
        verified: true,  // Google OAuth emails are pre-verified
        via: "email"
      }
    ],
  },
}
