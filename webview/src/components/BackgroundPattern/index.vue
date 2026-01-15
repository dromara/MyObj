<template>
  <div class="background-pattern" :class="patternClass">
    <!-- 网格背景 -->
    <div v-if="pattern === 'grid'" class="grid-pattern"></div>

    <!-- 点阵背景 -->
    <div v-else-if="pattern === 'dots'" class="dots-pattern"></div>

    <!-- 渐变背景 -->
    <div v-else-if="pattern === 'gradient'" class="gradient-pattern"></div>

    <!-- 波浪背景 -->
    <div v-else-if="pattern === 'waves'" class="waves-pattern">
      <svg viewBox="0 0 1200 120" preserveAspectRatio="none">
        <path d="M0,60 Q300,20 600,60 T1200,60 L1200,120 L0,120 Z" fill="currentColor" opacity="0.1"></path>
        <path d="M0,60 Q300,100 600,60 T1200,60 L1200,120 L0,120 Z" fill="currentColor" opacity="0.05"></path>
      </svg>
    </div>

    <!-- 粒子背景 -->
    <div v-else-if="pattern === 'particles'" class="particles-pattern"></div>
  </div>
</template>

<script setup lang="ts">
  type PatternType = 'none' | 'grid' | 'dots' | 'gradient' | 'waves' | 'particles'

  interface Props {
    pattern?: PatternType
  }

  const props = withDefaults(defineProps<Props>(), {
    pattern: 'none'
  })

  const patternClass = computed(() => {
    return `pattern-${props.pattern}`
  })
</script>

<style scoped>
  .background-pattern {
    position: fixed;
    inset: 0;
    pointer-events: none;
    z-index: 0;
    overflow: hidden;
  }

  /* 网格背景 */
  .grid-pattern {
    width: 100%;
    height: 100%;
    background-image:
      linear-gradient(rgba(37, 99, 235, 0.05) 1px, transparent 1px),
      linear-gradient(90deg, rgba(37, 99, 235, 0.05) 1px, transparent 1px);
    background-size: 50px 50px;
    opacity: 0.5;
  }

  html.dark .grid-pattern {
    background-image:
      linear-gradient(rgba(59, 130, 246, 0.1) 1px, transparent 1px),
      linear-gradient(90deg, rgba(59, 130, 246, 0.1) 1px, transparent 1px);
  }

  /* 点阵背景 */
  .dots-pattern {
    width: 100%;
    height: 100%;
    background-image: radial-gradient(circle, rgba(37, 99, 235, 0.1) 1px, transparent 1px);
    background-size: 30px 30px;
    opacity: 0.6;
  }

  html.dark .dots-pattern {
    background-image: radial-gradient(circle, rgba(59, 130, 246, 0.15) 1px, transparent 1px);
  }

  /* 渐变背景 */
  .gradient-pattern {
    width: 100%;
    height: 100%;
    background: linear-gradient(
      135deg,
      rgba(37, 99, 235, 0.05) 0%,
      rgba(79, 70, 229, 0.05) 50%,
      rgba(37, 99, 235, 0.05) 100%
    );
    background-size: 400% 400%;
    animation: gradient-flow 15s ease infinite;
  }

  html.dark .gradient-pattern {
    background: linear-gradient(
      135deg,
      rgba(59, 130, 246, 0.1) 0%,
      rgba(99, 102, 241, 0.1) 50%,
      rgba(59, 130, 246, 0.1) 100%
    );
  }

  /* 波浪背景 */
  .waves-pattern {
    width: 100%;
    height: 100%;
    position: absolute;
    bottom: 0;
    left: 0;
    color: var(--primary-color);
    animation: wave-animation 20s linear infinite;
  }

  @keyframes wave-animation {
    0% {
      transform: translateX(0);
    }
    100% {
      transform: translateX(-50%);
    }
  }

  .waves-pattern svg {
    width: 200%;
    height: 100%;
    display: block;
  }

  /* 粒子背景 */
  .particles-pattern {
    width: 100%;
    height: 100%;
    position: relative;
  }

  .particles-pattern::before {
    content: '';
    position: absolute;
    inset: 0;
    background-image:
      radial-gradient(2px 2px at 20% 30%, rgba(37, 99, 235, 0.3), transparent),
      radial-gradient(2px 2px at 60% 70%, rgba(79, 70, 229, 0.3), transparent),
      radial-gradient(1px 1px at 50% 50%, rgba(37, 99, 235, 0.2), transparent),
      radial-gradient(1px 1px at 80% 10%, rgba(79, 70, 229, 0.2), transparent),
      radial-gradient(2px 2px at 90% 60%, rgba(37, 99, 235, 0.2), transparent);
    background-size: 200% 200%;
    animation: particle-float 20s ease-in-out infinite;
  }

  html.dark .particles-pattern::before {
    background-image:
      radial-gradient(2px 2px at 20% 30%, rgba(59, 130, 246, 0.4), transparent),
      radial-gradient(2px 2px at 60% 70%, rgba(99, 102, 241, 0.4), transparent),
      radial-gradient(1px 1px at 50% 50%, rgba(59, 130, 246, 0.3), transparent),
      radial-gradient(1px 1px at 80% 10%, rgba(99, 102, 241, 0.3), transparent),
      radial-gradient(2px 2px at 90% 60%, rgba(59, 130, 246, 0.3), transparent);
  }

  @keyframes gradient-flow {
    0% {
      background-position: 0% 50%;
    }
    50% {
      background-position: 100% 50%;
    }
    100% {
      background-position: 0% 50%;
    }
  }

  @keyframes particle-float {
    0%,
    100% {
      transform: translate(0, 0) rotate(0deg);
    }
    25% {
      transform: translate(20px, -20px) rotate(90deg);
    }
    50% {
      transform: translate(-20px, 20px) rotate(180deg);
    }
    75% {
      transform: translate(20px, 20px) rotate(270deg);
    }
  }
</style>
