import { Injectable } from '@nestjs/common';

@Injectable()
export class AppService {
  async getHello(): Promise<string> {
    await new Promise((resolve) => setTimeout(resolve, 1000));
    return 'Hello World!';
  }
}
