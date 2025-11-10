import { Entity, Column, PrimaryColumn, OneToMany, CreateDateColumn, UpdateDateColumn } from 'typeorm';
import { ObjectType, Field, ID } from '@nestjs/graphql';
import { UserChannel } from './usersChannel.entity';

@ObjectType()
@Entity('users')
export class User {
  @Field(() => ID)
  @PrimaryColumn({ type: 'varchar', length: 50 })
  user_id: string;

  @Field()
  @Column({ type: 'varchar', length: 255, unique: true })
  email: string;

  @Field({ nullable: true })
  @Column({ type: 'varchar', length: 20, nullable: true })
  phone?: string;

  @Field()
  @Column({ type: 'varchar', length: 50, default: 'UTC' })
  timezone: string;

  @Field()
  @Column({ type: 'varchar', length: 10, default: 'en' })
  language: string;

  @Field()
  @Column({ type: 'boolean', default: true })
  notification_enabled: boolean;

  @Field()
  @Column({ type: 'boolean', default: false })
  marketing: boolean;

  @Field()
  @Column({ type: 'boolean', default: true })
  transactional: boolean;

  @Field()
  @Column({ type: 'boolean', default: true })
  reminders: boolean;

  @Field()
  @Column({ type: 'boolean', default: false })
  digest_enabled: boolean;

  @Field()
  @Column({ type: 'varchar', length: 20, default: 'daily' })
  digest_frequency: string;

  @Field()
  @Column({ type: 'varchar', length: 5, default: '09:00' })
  digest_time: string;

  @Field(() => [UserChannel], { nullable: true })
  @OneToMany(() => UserChannel, channel => channel.user, { cascade: true })
  channels: UserChannel[];

  @Field()
  @CreateDateColumn()
  created_at: Date;

  @Field()
  @UpdateDateColumn()
  updated_at: Date;
}
