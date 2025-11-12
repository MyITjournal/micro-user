import { Module } from '@nestjs/common';
import { ConfigModule, ConfigService } from '@nestjs/config';
import { GraphQLModule } from '@nestjs/graphql';
import { ApolloDriver, ApolloDriverConfig } from '@nestjs/apollo';
import { TypeOrmModule } from '@nestjs/typeorm';
import { AppController } from './app.controller';
import { AppService } from './app.service';
import { SimpleUsersModule } from './simple-users/simple-users.module';
import { ComplexUsersModule } from './complex-users/users.module';
import { HealthModule } from './health/health.module';
import { CacheModule } from './cache/cache.module';

@Module({
  imports: [
    // Config Module - Load environment variables
    ConfigModule.forRoot({
      isGlobal: true,
      envFilePath: '.env',
    }),

    // Cache Module - Redis configuration
    CacheModule,

    // GraphQL Configuration
    GraphQLModule.forRoot<ApolloDriverConfig>({
      driver: ApolloDriver,
      autoSchemaFile: true, // generates schema in memory
      playground: true, // enable GraphQL Playground
      path: '/api/v1/graphql', // GraphQL endpoint
    }),

    // TypeORM Configuration
    TypeOrmModule.forRootAsync({
      inject: [ConfigService],
      useFactory: (configService: ConfigService) => {
        const databaseUrl = configService.get<string>('DATABASE_URL');

        // Check if DATABASE_URL is provided (Heroku) & use it
        if (databaseUrl) {
          return {
            type: 'postgres',
            url: databaseUrl,
            entities: [__dirname + '/**/*.entity{.ts,.js}'],
            synchronize:
              configService.get('NODE_ENV') === 'production' ? false : true,
            logging: false,
            ssl:
              configService.get('NODE_ENV') === 'production'
                ? { rejectUnauthorized: false }
                : false,
          };
        }

        // Otherwise, use individual connection parameters
        return {
          type: 'postgres',
          host: configService.get<string>('DB_HOST', 'localhost'),
          port: configService.get<number>('DB_PORT', 5432),
          username: configService.get<string>('DB_USERNAME', 'postgres'),
          password: configService.get<string>('DB_PASSWORD'),
          database: configService.get<string>('DB_NAME', 'user_service'),
          entities: [__dirname + '/**/*.entity{.ts,.js}'],
          synchronize: true, // Auto-create tables on startup
          logging: false,
        };
      },
    }),

    SimpleUsersModule,
    ComplexUsersModule,
    HealthModule,
  ],
  controllers: [AppController],
  providers: [AppService],
})
export class AppModule {}
