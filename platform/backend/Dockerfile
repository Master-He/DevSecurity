ARG MODULE=core-service

FROM maven:3.9.9-eclipse-temurin-17 AS build
ARG MODULE
WORKDIR /workspace
COPY pom.xml .
COPY eureka-server/pom.xml eureka-server/pom.xml
COPY gateway-service/pom.xml gateway-service/pom.xml
COPY core-service/pom.xml core-service/pom.xml
RUN mvn -B -q -pl "${MODULE}" -am dependency:go-offline
COPY eureka-server eureka-server
COPY gateway-service gateway-service
COPY core-service core-service
RUN mvn -B -q -pl "${MODULE}" -am package -DskipTests

FROM eclipse-temurin:17-jre-jammy
RUN apt-get update \
    && apt-get install -y --no-install-recommends curl \
    && rm -rf /var/lib/apt/lists/* \
    && groupadd --system --gid 10001 app \
    && useradd --system --uid 10001 --gid app --home-dir /app app \
    && mkdir -p /var/lib/devsecurity/sca-work \
    && chown -R app:app /var/lib/devsecurity
ARG MODULE
WORKDIR /app
COPY --from=build "/workspace/${MODULE}/target/${MODULE}-0.1.0-SNAPSHOT.jar" /app/app.jar
USER app
EXPOSE 8080
ENTRYPOINT ["java", "-XX:MaxRAMPercentage=75", "-jar", "/app/app.jar"]
