# SouLink privacy overview

This document summarizes the product's privacy principles for contributors and
operators. The live product notice remains the user-facing source of truth.

## Data categories

SouLink may process:

- account information such as campus email and login credentials;
- profile and questionnaire responses used for matching;
- participation, result-view, connection, and feedback events required to
  operate and improve the service;
- basic browser, access-log, and technical information used for reliability and
  security.

## Purpose limitation

Data should be used only to:

- register, authenticate, and protect accounts;
- calculate and explain matches;
- provide a user-authorized path to continue a connection;
- deliver service notifications;
- operate, debug, secure, and improve the product;
- meet applicable legal obligations.

## Product boundaries

- Profiles are not a public browse feed.
- Contact information is not public by default.
- Result views should reveal only what is necessary for the matching experience.
- SouLink should not introduce public popularity rankings, continuous swipe
  exposure, or third-party advertising based on personal questionnaire data.

## Security practices

Contributors should preserve:

- TLS for production traffic;
- strong password hashing and authentication-secret rotation;
- least-privilege access to production systems;
- separation of source code from local/production secrets;
- logging that avoids questionnaire answers, credentials, and contact data;
- controlled exports and imports for matching runs;
- timely deletion, anonymization, or de-identification after account deletion,
  subject to legitimate security and legal retention requirements.

## Repository rules

Never commit:

- production `.env` files;
- database exports or questionnaire responses;
- SMTP or LLM API credentials;
- server addresses tied to private infrastructure;
- TLS certificates or private keys;
- personal contact lists or community invitation codes.

Use the committed `.env.example` files and keep real values in an approved
secret-management system.

## User rights and contact

Users can update their profile or request account and data deletion. Questions
about privacy, data handling, or deletion can be sent to
[thusoulink@163.com](mailto:thusoulink@163.com).

No online system can guarantee absolute security. The product should therefore
continue minimizing the sensitivity and volume of collected data.
